package server

import (
	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address string `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
}

func InitConfig() (*Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
