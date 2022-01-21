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
)

type Config struct {
	Address        string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
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
	command := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	reportInterval := command.Duration("r", DefaultReportInterval, "Report interval")
	pollInterval := command.Duration("p", DefaultPollInterval, "Poll interval")
	address := command.String("a", DefaultAddress, "Endpoint address")

	command.Parse(args)

	var cfg = Config{
		Address:        *address,
		ReportInterval: *reportInterval,
		PollInterval:   *pollInterval,
	}
	return &cfg
}
