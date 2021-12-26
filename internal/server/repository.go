package server

import (
	"fmt"

	"github.com/OmAsana/yapraktikum/internal/metrics"
)

type RepositoryError error

var (
	errorCounterNotFound RepositoryError = fmt.Errorf("Counter not found")
	errorGaugeNotFound   RepositoryError = fmt.Errorf("Gauge not found")
)

type MetricsRepository interface {
	StoreCounter(counter metrics.Counter) RepositoryError
	RetrieveCounter(name string) (metrics.Counter, RepositoryError)
	StoreGauge(gauge metrics.Gauge) RepositoryError
	RetrieveGauge(name string) (metrics.Gauge, RepositoryError)
}
