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

func startHTTPServer(wg *sync.WaitGroup) (*http.Server, error) {

	cfg, err := server.InitConfig()
	if err != nil {
		return nil, err
	}

	repo := server.NewRepositoryMock()
	handler, err := server.NewMetricsServer(
		repo,
		server.WithRestore(cfg.Restore),
		server.WithStoreFile(cfg.StoreFile),
		server.WithStoreInterval(cfg.StoreInterval),
	)
	if err != nil {
		return nil, err
	}
	srv := &http.Server{Addr: cfg.Address, Handler: handler}
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Println("Server shut down with err: ", err)
			handler.FlushToDisk()
		}
	}()
	return srv, nil

}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	waitServerShutdown := &sync.WaitGroup{}

	waitServerShutdown.Add(1)
	httpServer, err := startHTTPServer(waitServerShutdown)

	if err != nil {
		panic(err)
	}

	go func() {
		<-ctx.Done()
		httpServer.Shutdown(ctx)
	}()

	waitServerShutdown.Wait()

}
