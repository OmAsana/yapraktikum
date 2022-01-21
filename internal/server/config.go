package server

import (
	"time"

	"github.com/caarlos0/env/v6"
)

var (
	DefaultAddress       = "localhost:8080"
	DefaultStoreInterval = 300 * time.Second
	DefaultStoreFile     = "/tmp/devops-metrics-db.json"
	DefaultRestore       = true
)

type Config struct {
	//Address       string        `env:"ADDRESS" envDefault:"localhost:8080"`
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
}

func InitConfig() (*Config, error) {
	cfg := Config{
		Address:       DefaultAddress,
		StoreFile:     DefaultStoreFile,
		StoreInterval: DefaultStoreInterval,
		Restore:       DefaultRestore,
	}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
