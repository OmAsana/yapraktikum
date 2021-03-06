package repository

import (
	"fmt"

	"github.com/OmAsana/yapraktikum/internal/metrics"
)

type RepositoryError error

var (
	ErrorCounterNotFound  RepositoryError = fmt.Errorf("counter not found")
	ErrorCounterIsNoValid RepositoryError = fmt.Errorf("counter is not valid")
	ErrorGaugeNotFound    RepositoryError = fmt.Errorf("gauge not found")
	ErrorInternalError    RepositoryError = fmt.Errorf("internal error")
)

type MetricsRepository interface {
	StoreCounter(counter metrics.Counter) RepositoryError
	RetrieveCounter(name string) (metrics.Counter, RepositoryError)
	StoreGauge(gauge metrics.Gauge) RepositoryError
	RetrieveGauge(name string) (metrics.Gauge, RepositoryError)
	ListStoredMetrics() ([]metrics.Gauge, []metrics.Counter, RepositoryError)
	Ping() bool
	WriteBulkGauges(gauges []metrics.Gauge) error
	WriteBulkCounters(counters []metrics.Counter) error
}
