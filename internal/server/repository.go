package server

import (
	"fmt"

	"github.com/OmAsana/yapraktikum/internal/metrics"
)

type RepositoryError error

var (
	ErrorCounterNotFound RepositoryError = fmt.Errorf("counter not found")
	ErrorGaugeNotFound   RepositoryError = fmt.Errorf("gauge not found")
)

type MetricsRepository interface {
	StoreCounter(counter metrics.Counter) RepositoryError
	RetrieveCounter(name string) (metrics.Counter, RepositoryError)
	StoreGauge(gauge metrics.Gauge) RepositoryError
	RetrieveGauge(name string) (metrics.Gauge, RepositoryError)
}
