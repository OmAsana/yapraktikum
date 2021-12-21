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

	a := agent.NewDefaultAgent()
	a.Server(ctx)
	<-ctx.Done()
}
