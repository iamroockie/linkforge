package config

type RateLimitConfig struct {
	Rate  float64 `env:"RATE" envDefault:"0.1"`
	Burst int     `env:"BURST" envDefault:"10"`
}
