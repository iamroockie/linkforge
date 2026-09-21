package api

import (
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iamroockie/linkforge/internal/platform/config"
	"github.com/iamroockie/linkforge/internal/platform/httpx"
)

type Options struct {
	RateLimit      config.RateLimitConfig
	TrustedProxies []string
}

type App struct {
	PgxPool   *pgxpool.Pool
	Logger    *slog.Logger
	Link      LinkModule
	ClientIP  httpx.ClientIPResolver
	RateLimit RateLimitModule
}

func NewApp(pool *pgxpool.Pool, log *slog.Logger, opts Options) (App, error) {
	resolver, err := httpx.NewClientIPResolver(opts.TrustedProxies)
	if err != nil {
		return App{}, fmt.Errorf("create client IP resolver: %w", err)
	}
	rateLimit, err := newRateLimitModule(opts.RateLimit)
	if err != nil {
		return App{}, fmt.Errorf("create rate limit module: %w", err)
	}
	return App{
		PgxPool: pool,
		Logger:  log,

		Link:      newLinkModule(pool),
		ClientIP:  resolver,
		RateLimit: rateLimit,
	}, nil
}
