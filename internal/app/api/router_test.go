package api_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/app/api"
	"github.com/iamroockie/linkforge/internal/platform/config"
	"github.com/iamroockie/linkforge/internal/platform/httpx/httpxtest"
	"github.com/iamroockie/linkforge/internal/ratelimit"
	"github.com/iamroockie/linkforge/internal/ratelimit/adapter/memory"
)

type limiterFunc func(context.Context, string) (ratelimit.Decision, error)

func (f limiterFunc) Allow(ctx context.Context, key string) (ratelimit.Decision, error) {
	return f(ctx, key)
}

func TestRouter_RateLimitCreateLink(t *testing.T) {
	app, err := api.NewApp(nil, slog.New(slog.DiscardHandler), api.Options{
		RateLimit:      config.RateLimitConfig{Rate: 1, Burst: 1},
		TrustedProxies: []string{"10.0.0.0/8"},
	})
	require.NoError(t, err)
	now := time.Now()
	app.RateLimit.Limiter, err = memory.NewTokenBucket(1, 1, func() time.Time { return now })
	require.NoError(t, err)
	router := api.NewRouter(app)
	request := func(ip string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/links", nil)
		req.RemoteAddr = "10.0.0.1:80"
		req.Header.Set("X-Forwarded-For", ip)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	// Invalid bodies spend quota without reaching the database.
	first := request("192.0.2.1")
	require.Equal(t, http.StatusUnsupportedMediaType, first.Code)
	denied := request("192.0.2.1")
	require.Equal(t, http.StatusTooManyRequests, denied.Code)
	assert.Equal(t, "1", denied.Header().Get("Retry-After"))
	httpxtest.CheckResponseError(t, denied, "too_many_requests", "Too many requests")
	spoofed := request("garbage, 192.0.2.1")
	require.Equal(t, http.StatusTooManyRequests, spoofed.Code)
	malformed := request("192.0.2.1, garbage")
	require.Equal(t, http.StatusBadRequest, malformed.Code)
	httpxtest.CheckResponseError(t, malformed, "invalid_request", "Invalid request")
	otherClient := request("192.0.2.2")
	require.Equal(t, http.StatusUnsupportedMediaType, otherClient.Code)
	now = now.Add(time.Second)
	restored := request("192.0.2.1")
	require.Equal(t, http.StatusUnsupportedMediaType, restored.Code)
}

func TestRouter_RateLimitScope(t *testing.T) {
	// A closed pool makes DB-backed handlers fail without connecting to a database.
	pool, err := pgxpool.New(t.Context(), "postgres://localhost/test")
	require.NoError(t, err)
	pool.Close()
	app, err := api.NewApp(pool, slog.New(slog.DiscardHandler), api.Options{
		RateLimit: config.RateLimitConfig{Rate: 1, Burst: 10},
	})
	require.NoError(t, err)
	checks := 0
	app.RateLimit.Limiter = limiterFunc(func(
		_ context.Context, key string,
	) (ratelimit.Decision, error) {
		checks++
		assert.Equal(t, "create_link:192.0.2.1", key)
		return ratelimit.Decision{RetryAfter: time.Second}, nil
	})
	router := api.NewRouter(app)
	tests := map[string]struct {
		method     string
		path       string
		wantStatus int
		wantChecks int
	}{
		"create link": {
			method:     http.MethodPost,
			path:       "/links",
			wantStatus: http.StatusTooManyRequests,
			wantChecks: 1,
		},
		"health": {
			method:     http.MethodGet,
			path:       "/healthz",
			wantStatus: http.StatusOK,
		},
		"readiness": {
			method:     http.MethodGet,
			path:       "/readyz",
			wantStatus: http.StatusServiceUnavailable,
		},
		"redirect": {
			method:     http.MethodGet,
			path:       "/r/example",
			wantStatus: http.StatusInternalServerError,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			checks = 0
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.RemoteAddr = "192.0.2.1:80"
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantChecks, checks)
		})
	}
}
