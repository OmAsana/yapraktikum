package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"sync"
	"syscall"

	"github.com/OmAsana/yapraktikum/internal/logging"
	"github.com/OmAsana/yapraktikum/internal/repository"
	"github.com/OmAsana/yapraktikum/internal/repository/inmemory_store"
	"github.com/OmAsana/yapraktikum/internal/repository/sql"
	"github.com/OmAsana/yapraktikum/internal/server"
)

func startHTTPServer(wg *sync.WaitGroup) (*http.Server, error) {

	cfg, err := server.InitConfig()
	if err != nil {
		return nil, err
	}
	logging.Log.S().Infof("Config: %+v", cfg)

	var repo repository.MetricsRepository
	if cfg.DatabaseDSN != "" {
		repo, err = sql.NewRepository(cfg.DatabaseDSN, cfg.Restore)

	} else {
		repo, err = inmemory_store.NewInMemoryRepo(
			inmemory_store.WithRestore(cfg.Restore),
			inmemory_store.WithStoreFile(cfg.StoreFile),
			inmemory_store.WithStoreInterval(cfg.StoreInterval),
		)
	}
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
