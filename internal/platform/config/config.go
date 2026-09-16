package config

import (
	"log/slog"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Env             Env            `env:"ENV" envDefault:"dev"`
	HTTPAddr        string         `env:"HTTP_ADDR" envDefault:":8080"`
	LogLevel        slog.Level     `env:"LOG_LEVEL" envDefault:"info"`
	ShutdownTimeout time.Duration  `env:"SHUTDOWN_TIMEOUT" envDefault:"20s"`
	Postgres        PostgresConfig `envPrefix:"PG_"`
}

func Load() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
