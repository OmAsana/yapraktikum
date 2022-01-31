package agent

import (
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

var (
	DefaultAddress        = "127.0.0.1:8080"
	DefaultReportInterval = 10 * time.Second
	DefaultPollInterval   = 2 * time.Second
	DefaultHashKey        = ""
	DefaultLogLevel       = "info"

	DefaultConfig = Config{
		Address:        DefaultAddress,
		ReportInterval: DefaultReportInterval,
		PollInterval:   DefaultPollInterval,
		HaskKey:        DefaultHashKey,
		LogLevel:       DefaultLogLevel,
	}
)

type Config struct {
	Address        string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	HaskKey        string        `env:"KEY"`
	LogLevel       string        `env:"LOG_LEVEL"`
	command        *flag.FlagSet
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

func (c *Config) initEnvArgs() error {
	if err := env.Parse(c); err != nil {
		return err
	}
	return nil
}

func (c *Config) initCmdFlagsWithArgs(args []string) error {
	command := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	reportInterval := command.Duration("r", DefaultReportInterval, "Report interval")
	pollInterval := command.Duration("p", DefaultPollInterval, "Poll interval")
	address := command.String("a", DefaultAddress, "Endpoint address")
	hashKey := command.String("k", DefaultHashKey, "Hash key")
	logLevel := command.String("log_level", DefaultLogLevel, "Log level")

	if err := command.Parse(args); err != nil {
		return err
	}

	c.ReportInterval = *reportInterval
	c.PollInterval = *pollInterval
	c.Address = *address
	c.HaskKey = *hashKey
	c.LogLevel = *logLevel

	return nil
}
