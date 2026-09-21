package rest_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/platform/httpx"
	"github.com/iamroockie/linkforge/internal/platform/httpx/httpxtest"
	"github.com/iamroockie/linkforge/internal/ratelimit"
	"github.com/iamroockie/linkforge/internal/ratelimit/transport/rest"
)

type limiterFunc func(context.Context, string) (ratelimit.Decision, error)

func (f limiterFunc) Allow(ctx context.Context, key string) (ratelimit.Decision, error) {
	return f(ctx, key)
}

func TestRateLimit_Allowed(t *testing.T) {
	ctx := httpx.ReportErrorCtx(t.Context())
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/links", nil)
	req.RemoteAddr = "192.0.2.1:80"
	checks, calls := 0, 0
	limiter := limiterFunc(func(got context.Context, key string) (ratelimit.Decision, error) {
		checks++
		assert.Equal(t, ctx, got)
		assert.Equal(t, "create_link:192.0.2.1", key)
		return ratelimit.Decision{Allowed: true}, nil
	})
	next := http.HandlerFunc(func(w http.ResponseWriter, got *http.Request) {
		calls++
		assert.Same(t, req, got)
		w.Header().Set("X-Handler", "called")
		w.WriteHeader(http.StatusAccepted)
		_, err := w.Write([]byte("accepted"))
		assert.NoError(t, err)
	})
	handler := rest.RateLimit(limiter, httpx.ClientIPResolver{}, "create_link")(next)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, 1, checks)
	assert.Equal(t, 1, calls)
	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.Equal(t, "called", rec.Header().Get("X-Handler"))
	assert.Equal(t, "accepted", rec.Body.String())
	assert.Empty(t, rec.Header().Get("Retry-After"))
	require.NoError(t, httpx.GetReportedError(ctx))
}

func TestRateLimit_ClientKey(t *testing.T) {
	tests := map[string]struct {
		prefix    string
		peer      string
		forwarded string
		trusted   []string
		wantKey   string
	}{
		"custom prefix": {
			prefix: "redirect", peer: "192.0.2.1:80", wantKey: "redirect:192.0.2.1",
		},
		"trusted proxy": {
			prefix: "create_link", peer: "10.0.0.1:80", forwarded: "192.0.2.1",
			trusted: []string{"10.0.0.0/8"}, wantKey: "create_link:192.0.2.1",
		},
		"untrusted forwarded header": {
			prefix: "create_link", peer: "192.0.2.1:80", forwarded: "203.0.113.1",
			wantKey: "create_link:192.0.2.1",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resolver, err := httpx.NewClientIPResolver(tt.trusted)
			require.NoError(t, err)
			checks := 0
			limiter := limiterFunc(func(_ context.Context, key string) (ratelimit.Decision, error) {
				checks++
				assert.Equal(t, tt.wantKey, key)
				return ratelimit.Decision{Allowed: true}, nil
			})
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})
			req := httptest.NewRequest(http.MethodPost, "/links", nil)
			req.RemoteAddr = tt.peer
			req.Header.Set("X-Forwarded-For", tt.forwarded)
			rec := httptest.NewRecorder()

			rest.RateLimit(limiter, resolver, tt.prefix)(next).ServeHTTP(rec, req)

			assert.Equal(t, 1, checks)
			assert.Equal(t, http.StatusNoContent, rec.Code)
		})
	}
}

func TestRateLimit_Denied(t *testing.T) {
	tests := map[string]struct {
		retryAfter time.Duration
		wantHeader string
	}{
		"zero":              {retryAfter: 0, wantHeader: "1"},
		"subsecond":         {retryAfter: time.Nanosecond, wantHeader: "1"},
		"exact second":      {retryAfter: time.Second, wantHeader: "1"},
		"fractional second": {retryAfter: 1500 * time.Millisecond, wantHeader: "2"},
		"just over second":  {retryAfter: time.Second + time.Nanosecond, wantHeader: "2"},
		"multiple seconds":  {retryAfter: 3 * time.Second, wantHeader: "3"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := httpx.ReportErrorCtx(t.Context())
			checks := 0
			limiter := limiterFunc(func(context.Context, string) (ratelimit.Decision, error) {
				checks++
				return ratelimit.Decision{RetryAfter: tt.retryAfter}, nil
			})
			next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Fatal("denied request reached the next handler")
			})
			handler := rest.RateLimit(limiter, httpx.ClientIPResolver{}, "create_link")(next)
			req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/links", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, 1, checks)
			require.Equal(t, http.StatusTooManyRequests, rec.Code)
			assert.Equal(t, tt.wantHeader, rec.Header().Get("Retry-After"))
			httpxtest.CheckResponseError(t, rec, "too_many_requests", "Too many requests")
			require.NoError(t, httpx.GetReportedError(ctx))
		})
	}
}

func TestRateLimit_LimiterError(t *testing.T) {
	failure := errors.New("backend unavailable")
	ctx := httpx.ReportErrorCtx(t.Context())
	checks := 0
	limiter := limiterFunc(func(context.Context, string) (ratelimit.Decision, error) {
		checks++
		return ratelimit.Decision{}, failure
	})
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("request reached the next handler after limiter error")
	})
	handler := rest.RateLimit(limiter, httpx.ClientIPResolver{}, "create_link")(next)
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/links", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, 1, checks)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Empty(t, rec.Header().Get("Retry-After"))
	httpxtest.CheckResponseError(t, rec, "internal_error", "Internal server error")
	require.ErrorIs(t, httpx.GetReportedError(ctx), failure)
}

func TestRateLimit_InvalidRemote(t *testing.T) {
	ctx := httpx.ReportErrorCtx(t.Context())
	limiter := limiterFunc(func(context.Context, string) (ratelimit.Decision, error) {
		t.Fatal("invalid remote address reached the limiter")
		return ratelimit.Decision{}, nil
	})
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("invalid remote address reached the next handler")
	})
	handler := rest.RateLimit(limiter, httpx.ClientIPResolver{}, "create_link")(next)
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/links", nil)
	req.RemoteAddr = "invalid"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Empty(t, rec.Header().Get("Retry-After"))
	httpxtest.CheckResponseError(t, rec, "invalid_request", "Invalid request")
	require.NoError(t, httpx.GetReportedError(ctx))
}

func TestRateLimit_InvalidForwardedChain(t *testing.T) {
	resolver, err := httpx.NewClientIPResolver([]string{"10.0.0.0/8"})
	require.NoError(t, err)
	limiter := limiterFunc(func(context.Context, string) (ratelimit.Decision, error) {
		t.Fatal("invalid forwarded chain reached the limiter")
		return ratelimit.Decision{}, nil
	})
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("invalid forwarded chain reached the next handler")
	})
	handler := rest.RateLimit(limiter, resolver, "create_link")(next)
	req := httptest.NewRequest(http.MethodPost, "/links", nil)
	req.RemoteAddr = "10.0.0.1:80"
	req.Header.Set("X-Forwarded-For", "192.0.2.1, garbage")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Empty(t, rec.Header().Get("Retry-After"))
	httpxtest.CheckResponseError(t, rec, "invalid_request", "Invalid request")
}
