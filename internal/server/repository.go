package server

import (
	"fmt"

	"github.com/OmAsana/yapraktikum/internal/metrics"
)

type RepositoryError error

var (
	ErrorCounterNotFound RepositoryError = fmt.Errorf("Counter not found")
	ErrorGaugeNotFound   RepositoryError = fmt.Errorf("Gauge not found")
)

type MetricsRepository interface {
	StoreCounter(counter metrics.Counter) error
	RetrieveCounter(name string) (metrics.Counter, error)
	StoreGauge(gauge metrics.Gauge) error
	RetrieveGauge(name string) (metrics.Gauge, error)
}
