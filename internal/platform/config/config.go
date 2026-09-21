package config

import (
	"log/slog"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Env             Env             `env:"ENV" envDefault:"dev"`
	LogLevel        slog.Level      `env:"LOG_LEVEL" envDefault:"info"`
	HTTP            HTTPConfig      `envPrefix:"HTTP_"`
	TrustedProxies  []string        `env:"TRUSTED_PROXIES"`
	ShutdownTimeout time.Duration   `env:"SHUTDOWN_TIMEOUT" envDefault:"20s"`
	RateLimit       RateLimitConfig `envPrefix:"RATELIMIT_"`
	Postgres        PostgresConfig  `envPrefix:"PG_"`
}

func Load() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
