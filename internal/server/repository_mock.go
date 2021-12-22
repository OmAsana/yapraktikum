package server

import (
	"github.com/OmAsana/yapraktikum/internal/metrics"
)

type RepositoryMock struct {
	gauges   map[string]float64
	counters map[string]int64
}

func (r *RepositoryMock) RetrieveCounter(name string) (metrics.Counter, RepositoryError) {
	if v, ok := r.counters[name]; ok {
		return metrics.Counter{
			Name:  name,
			Value: v,
		}, nil
	}
	return metrics.Counter{}, ErrorCounterNotFound
}

func (r *RepositoryMock) RetrieveGauge(name string) (metrics.Gauge, RepositoryError) {
	if v, ok := r.gauges[name]; ok {
		return metrics.Gauge{
			Name:  name,
			Value: v,
		}, nil
	}
	return metrics.Gauge{}, ErrorGaugeNotFound
}

func (r *RepositoryMock) StoreCounter(counter metrics.Counter) RepositoryError {
	_, ok := r.counters[counter.Name]
	if ok {
		r.counters[counter.Name] += counter.Value
		return nil
	}
	r.counters[counter.Name] = counter.Value

	return nil
}

func (r *RepositoryMock) StoreGauge(gauge metrics.Gauge) RepositoryError {
	r.gauges[gauge.Name] = gauge.Value
	return nil
}
