package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/OmAsana/yapraktikum/internal/agent"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	cfg, err := agent.InitConfig()
	if err != nil {
		panic(err)
	}
	a, err := agent.NewAgentWithOptions(
		agent.WithAddress(cfg.Address),
		agent.WithPollInterval(cfg.PollInterval),
		agent.WithReportInterval(cfg.ReportInterval),
	)
	if err != nil {
		panic(err)
	}
	a.Server(ctx)
	<-ctx.Done()
}
