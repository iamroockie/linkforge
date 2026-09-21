package config

import (
	"net"
	"strconv"
)

type HTTPConfig struct {
	Host string `env:"HOST" envDefault:"0.0.0.0"`
	Port uint16 `env:"PORT" envDefault:"8080"`
}

func (c HTTPConfig) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(int(c.Port)))
}
