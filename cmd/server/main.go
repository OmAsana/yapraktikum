package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/OmAsana/yapraktikum/internal/logging"
	"github.com/OmAsana/yapraktikum/internal/repository"
	"github.com/OmAsana/yapraktikum/internal/repository/inmemoryStore"
	"github.com/OmAsana/yapraktikum/internal/repository/sql"
	"github.com/OmAsana/yapraktikum/internal/server"
)

func startHTTPServer() (*http.Server, error) {

	cfg, err := server.InitConfig()
	if err != nil {
		return nil, err
	}
	logging.Log.S().Infof("Config: %+v", cfg)

	var repo repository.MetricsRepository
	if cfg.DatabaseDSN != "" {
		repo, err = sql.NewRepository(cfg.DatabaseDSN, cfg.Restore)

	} else {
		repo, err = inmemoryStore.NewInMemoryRepo(
			inmemoryStore.WithRestore(cfg.Restore),
			inmemoryStore.WithStoreFile(cfg.StoreFile),
			inmemoryStore.WithStoreInterval(cfg.StoreInterval),
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
		defer handler.FlushToDisk()

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logging.Log.S().Error("Server shut down with err: ", err)
		}
	}()
	return srv, nil

}

func main() {
	sigGracefullQuit, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	httpServer, err := startHTTPServer()

	if err != nil {
		panic(err)
	}

	connectionsClosed := make(chan struct{})
	go func() {
		<-sigGracefullQuit.Done()
		if err := httpServer.Shutdown(context.Background()); err != nil {
			logging.Log.S().Errorf("Error on shutdown: %s", err)
		}

		close(connectionsClosed)

	}()

	<-connectionsClosed
	defer logging.Log.Flush()

}
