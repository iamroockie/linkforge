package api

import (
	"time"

	"github.com/iamroockie/linkforge/internal/platform/config"
	"github.com/iamroockie/linkforge/internal/ratelimit/adapter/memory"
	"github.com/iamroockie/linkforge/internal/ratelimit/transport/rest"
)

type RateLimitModule struct {
	Limiter rest.Limiter
}

func newRateLimitModule(cfg config.RateLimitConfig) (RateLimitModule, error) {
	limiter, err := memory.NewTokenBucket(cfg.Rate, cfg.Burst, time.Now)
	if err != nil {
		return RateLimitModule{}, err
	}

	return RateLimitModule{Limiter: limiter}, nil
}
