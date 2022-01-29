package server

import (
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/OmAsana/yapraktikum/internal/logging"
)

var (
	DefaultAddress       = "localhost:8080"
	DefaultStoreInterval = 300 * time.Second
	DefaultStoreFile     = "/tmp/devops-metrics-db.json"
	DefaultRestore       = true
	DefaultHashKey       = ""
	DefaultDatabaseDSN   = ""

	DefaultConfig = Config{
		Address:       DefaultAddress,
		StoreInterval: DefaultStoreInterval,
		StoreFile:     DefaultStoreFile,
		Restore:       DefaultRestore,
		DatabaseDSN:   DefaultDatabaseDSN,
	}
)

type Config struct {
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	HashKey       string        `env:"KEY"`
	DatabaseDSN   string        `env:"DATABASE_DSN"`
}

func InitConfig() (*Config, error) {
	cfg := DefaultConfig

	if err := cfg.initCmdFlagsWithArgs(os.Args[1:]); err != nil {
		return nil, err
	}

	if err := cfg.initEnvArgs(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) initCmdFlagsWithArgs(args []string) error {
	command := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	address := command.String("a", DefaultAddress, "Listen on address")
	restore := command.Bool("r", DefaultRestore, "Restore metrics on startup")
	storeInterval := command.Duration("i", DefaultStoreInterval, "Store interval")
	storeFile := command.String("f", DefaultStoreFile, "Store file")
	hashKey := command.String("k", DefaultHashKey, "Hash key")
	databaseDSN := command.String("d", DefaultDatabaseDSN, "Postgre database connection string")

	if err := command.Parse(args); err != nil {
		return err
	}

	c.Address = *address
	c.Restore = *restore
	c.StoreFile = *storeFile
	c.StoreInterval = *storeInterval
	c.HashKey = *hashKey
	c.DatabaseDSN = *databaseDSN
	logging.Log.S().Infof("Config from flags: %+v", c)

	return nil
}

func (c *Config) initEnvArgs() error {
	if err := env.Parse(c); err != nil {
		return err
	}
	logging.Log.S().Infof("Config after envs vars: %+v", c)
	return nil
}
