package agent

import (
	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address        string `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	ReportInterval int64  `env:"REPORT_INTERVAL" envDefault:"2"`
	PollInterval   int64  `env:"POLL_INTERVAL" envDefault:"2"`
}

func InitConfig() (*Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
