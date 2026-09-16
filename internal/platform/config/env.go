package config

import "fmt"

type Env string

const (
	EnvProd Env = "prod"
	EnvDev  Env = "dev"
)

func (e *Env) UnmarshalText(text []byte) error {
	str := Env(text)
	switch str {
	case EnvProd, EnvDev:
		*e = Env(str)
		return nil
	default:
		return fmt.Errorf("invalid env: %s", str)
	}
}
