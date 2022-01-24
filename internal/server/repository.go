package server

import (
	"fmt"

	"github.com/OmAsana/yapraktikum/internal/metrics"
)

type RepositoryError error

var (
	ErrorCounterNotFound  RepositoryError = fmt.Errorf("counter not found")
	ErrorCounterIsNoValid RepositoryError = fmt.Errorf("counter is not valid")
	ErrorGaugeNotFound    RepositoryError = fmt.Errorf("gauge not found")
)

type MetricsRepository interface {
	StoreCounter(counter metrics.Counter) RepositoryError
	RetrieveCounter(name string) (metrics.Counter, RepositoryError)
	StoreGauge(gauge metrics.Gauge) RepositoryError
	RetrieveGauge(name string) (metrics.Gauge, RepositoryError)
	ListStoredMetrics() ([]metrics.Gauge, []metrics.Counter, RepositoryError)
}
