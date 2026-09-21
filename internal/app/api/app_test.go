package api_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/app/api"
	"github.com/iamroockie/linkforge/internal/platform/config"
	"github.com/iamroockie/linkforge/internal/platform/pg/pgtest"
)

func TestNewApp(t *testing.T) {
	pool := pgtest.NewPool(t)

	app, err := api.NewApp(pool, slog.New(slog.DiscardHandler), api.Options{
		RateLimit: config.RateLimitConfig{Rate: 1, Burst: 10},
	})
	require.NoError(t, err)

	assert.NotNil(t, app.PgxPool)
	assert.NotNil(t, app.Logger)
	assert.NotZero(t, app.Link.CreateLinkUC)
	assert.NotZero(t, app.Link.GetRedirectURLUC)
	assert.NotNil(t, app.RateLimit.Limiter)
}

func TestNewApp_InvalidOptions(t *testing.T) {
	tests := map[string]struct {
		opts      api.Options
		wantError string
	}{
		"invalid proxy": {
			opts: api.Options{
				RateLimit:      config.RateLimitConfig{Rate: 1, Burst: 10},
				TrustedProxies: []string{"invalid"},
			},
			wantError: "create client IP resolver",
		},
		"invalid rate": {
			opts:      api.Options{RateLimit: config.RateLimitConfig{Rate: 0, Burst: 10}},
			wantError: "create rate limit module",
		},
		"invalid burst": {
			opts:      api.Options{RateLimit: config.RateLimitConfig{Rate: 1, Burst: 0}},
			wantError: "create rate limit module",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := api.NewApp(nil, slog.New(slog.DiscardHandler), tt.opts)
			require.ErrorContains(t, err, tt.wantError)
		})
	}
}
