package sql

import "github.com/OmAsana/yapraktikum/internal/metrics"

type Counter struct {
	Name  string
	Value int64
}

func (c Counter) ToMetric() metrics.Counter {
	return metrics.Counter{
		Name:  c.Name,
		Value: c.Value,
	}
}

type Gauge struct {
	Name  string
	Delta float64
}

func (g Gauge) ToMetric() metrics.Gauge {
	return metrics.Gauge{
		Name:  g.Name,
		Value: g.Delta,
	}
}
