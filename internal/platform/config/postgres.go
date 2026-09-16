package config

import (
	"net"
	"net/url"
	"strconv"
)

type PostgresConfig struct {
	Host     string `env:"HOST,notEmpty"`
	Port     uint16 `env:"PORT,notEmpty"`
	DB       string `env:"DB,notEmpty"`
	User     string `env:"USER,notEmpty"`
	Password string `env:"PASSWORD,notEmpty"`
	SSLMode  string `env:"SSLMODE,notEmpty"`
}

func (c PostgresConfig) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(int(c.Port))),
		Path:   "/" + c.DB,
	}
	query := url.Values{}
	query.Set("sslmode", c.SSLMode)
	u.RawQuery = query.Encode()

	return u.String()
}
