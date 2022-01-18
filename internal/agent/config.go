package agent

import (
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address        string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
}

func InitConfig() (*Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
