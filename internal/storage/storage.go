package storage

import (
	"time"

	"github.com/OmAsana/yapraktikum/internal/metrics"
	"github.com/OmAsana/yapraktikum/internal/server"
)

var _ server.MetricsRepository = &CachedDatabaseWriter{}

type CachedDatabaseWriter struct {
	db            server.MetricsRepository
	storeInterval time.Duration
	storeFile     string
	restore       bool
}

func (d *CachedDatabaseWriter) StoreCounter(counter metrics.Counter) server.RepositoryError {
	return d.db.StoreCounter(counter)
}

func (d *CachedDatabaseWriter) RetrieveCounter(name string) (metrics.Counter, server.RepositoryError) {
	return d.db.RetrieveCounter(name)
}

func (d *CachedDatabaseWriter) StoreGauge(gauge metrics.Gauge) server.RepositoryError {
	return d.db.StoreGauge(gauge)
}

func (d *CachedDatabaseWriter) RetrieveGauge(name string) (metrics.Gauge, server.RepositoryError) {
	return d.db.RetrieveGauge(name)
}

func (d *CachedDatabaseWriter) ListStoredMetrics() ([]metrics.Gauge, []metrics.Counter, server.RepositoryError) {
	return d.db.ListStoredMetrics()
}

type DatabaseWriterOptions func(d *CachedDatabaseWriter) error

func (d *CachedDatabaseWriter) NewCachedDatabaseWriter(db server.MetricsRepository, opts ...DatabaseWriterOptions) (*CachedDatabaseWriter, error) {
	defaultDatabaseWriter := CachedDatabaseWriter{
		db:            db,
		storeInterval: 300 * time.Second,
		storeFile:     "/tmp/devops-metrics-db.json",
		restore:       true,
	}

	for _, opt := range opts {
		err := opt(&defaultDatabaseWriter)
		if err != nil {
			return nil, err
		}
	}

	return &defaultDatabaseWriter, nil
}
