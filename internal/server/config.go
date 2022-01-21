package server

import (
	"flag"
	"fmt"
	"os"
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
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
}

func InitConfig() (*Config, error) {
	cfg := initCmdFlags()
	return initEnvArgs(*cfg)
}

func initEnvArgs(cfg Config) (*Config, error) {
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func initCmdFlags() *Config {
	return initCmdFlagsWithArgs(os.Args[1:])
}

func initCmdFlagsWithArgs(args []string) *Config {
	fmt.Println(args)
	command := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	address := command.String("a", DefaultAddress, "Listen on address")
	restore := command.Bool("r", DefaultRestore, "Restore metrics on startup")
	storeInterval := command.Duration("i", DefaultStoreInterval, "Store interval")
	storeFile := command.String("f", DefaultStoreFile, "Store file")

	command.Parse(args)

	var cfg = Config{
		Address:       *address,
		StoreInterval: *storeInterval,
		StoreFile:     *storeFile,
		Restore:       *restore,
	}
	return &cfg
}
