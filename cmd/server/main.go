package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"sync"
	"syscall"

	"github.com/OmAsana/yapraktikum/internal/repository/mock"
	"github.com/OmAsana/yapraktikum/internal/server"
)

func startHTTPServer(wg *sync.WaitGroup) (*http.Server, error) {

	cfg, err := server.InitConfig()
	if err != nil {
		return nil, err
	}

	repo, err := mock.NewInMemoryRepo(
		mock.WithRestore(cfg.Restore),
		mock.WithStoreFile(cfg.StoreFile),
		mock.WithStoreInterval(cfg.StoreInterval),
	)
	if err != nil {
		return nil, err
	}
	handler, err := server.NewMetricsServer(
		repo,
		server.WithHashKey(cfg.HashKey),
	)
	if err != nil {
		return nil, err
	}
	srv := &http.Server{Addr: cfg.Address, Handler: handler}
	go func() {
		defer wg.Done()
		defer handler.FlushToDisk()

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Println("Server shut down with err: ", err)
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
