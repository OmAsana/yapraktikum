package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/OmAsana/yapraktikum/internal/logging"
	"github.com/OmAsana/yapraktikum/internal/repository"
	"github.com/OmAsana/yapraktikum/internal/repository/inmemorystore"
	"github.com/OmAsana/yapraktikum/internal/repository/sql"
	"github.com/OmAsana/yapraktikum/internal/server"
)

func startHTTPServer(addr string, handler *server.MetricsServer, logger *logging.Logger) (*http.Server, error) {

	srv := &http.Server{Addr: addr, Handler: handler}
	go func() {
		defer handler.FlushToDisk()

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.S().Error("Server shut down with err: ", err)
		}
	}()
	return srv, nil
}

func main() {

	sigGracefullQuit, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	logger := logging.NewLogger()

	cfg, err := server.InitConfig()
	if err != nil {
		logger.S().Panic("Could not init config: %s", err)
	}

	if err := logger.SetLogLevel(cfg.LogLevel); err != nil {
		logger.S().Panic(err)
	}

	repo, err := setupRepo(cfg, logger)
	if err != nil {
		logger.S().Panic("Could not init config: %s", err)
	}

	handler, err := setupHandler(repo, cfg, logger)
	if err != nil {
		logger.S().Panic("Could not setup handler: %s", err)
	}

	httpServer, err := startHTTPServer(cfg.Address, handler, logger)

	if err != nil {
		logger.S().Panic(err)
	}

	connectionsClosed := make(chan struct{})
	go func() {
		<-sigGracefullQuit.Done()
		if err := httpServer.Shutdown(context.Background()); err != nil {
			logger.S().Errorf("Error on shutdown: %s", err)
		}

		close(connectionsClosed)

	}()

	<-connectionsClosed
	defer logger.Flush()

}

func setupHandler(repo repository.MetricsRepository, cfg *server.Config, logger *logging.Logger) (*server.MetricsServer, error) {
	handler, err := server.NewMetricsServer(
		repo,
		server.WithHashKey(cfg.HashKey),
		server.WithLogger(logger),
	)
	return handler, err
}

func setupRepo(cfg *server.Config, logger *logging.Logger) (repository.MetricsRepository, error) {
	var repo repository.MetricsRepository
	var err error
	if cfg.DatabaseDSN != "" {
		repo, err = sql.NewRepository(cfg.DatabaseDSN, cfg.Restore, sql.WithLogger(logger))

	} else {
		repo, err = inmemorystore.NewInMemoryRepo(
			inmemorystore.WithRestore(cfg.Restore),
			inmemorystore.WithStoreFile(cfg.StoreFile),
			inmemorystore.WithStoreInterval(cfg.StoreInterval),
			inmemorystore.WithLogger(logger),
		)
	}
	return repo, err
}
