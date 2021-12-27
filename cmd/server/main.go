package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"sync"
	"syscall"

	"github.com/OmAsana/yapraktikum/internal/server"
)

func startHTTPServer(wg *sync.WaitGroup, handler http.Handler) *http.Server {
	srv := &http.Server{Addr: "127.0.0.1:8080", Handler: handler}
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Println("Server shut down with err: ", err)
		}
	}()
	return srv

}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	repo := server.NewRepositoryMock()
	metricsServer := server.NewMetricsServer(repo)

	waitServerShutdown := &sync.WaitGroup{}

	waitServerShutdown.Add(1)
	httpServer := startHTTPServer(waitServerShutdown, metricsServer)

	go func() {
		<-ctx.Done()
		httpServer.Shutdown(ctx)
	}()

	waitServerShutdown.Wait()

}
