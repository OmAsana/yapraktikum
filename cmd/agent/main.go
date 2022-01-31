package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/OmAsana/yapraktikum/internal/agent"
	"github.com/OmAsana/yapraktikum/internal/logging"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	logger := logging.NewLogger()

	cfg, err := agent.InitConfig()
	if err != nil {
		logger.S().Panic(err)
	}

	a, err := agent.NewAgentWithOptions(
		agent.WithAddress(cfg.Address),
		agent.WithPollInterval(cfg.PollInterval),
		agent.WithReportInterval(cfg.ReportInterval),
		agent.WithLogger(logger),
		agent.WithHashKey(cfg.HaskKey),
	)

	if err != nil {
		panic(err)
	}
	a.Server(ctx)
	<-ctx.Done()
}
