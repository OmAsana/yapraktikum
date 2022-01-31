package metrics

import (
	"math/rand"
)

type Registry struct {
	Gauges      []Gauge
	Counters    []Counter
	PollCounter Counter
}

func NewRegistry() *Registry {
	return &Registry{PollCounter: Counter{
		Name:  "PollCount",
		Value: 0,
	}}
}

func (r *Registry) Collect() error {
	var err error
	r.Gauges, err = CollectRuntimeMetrics()
	if err != nil {
		return err
	}

	r.Gauges = append(r.Gauges, func() Gauge {
		return Gauge{
			Name:  "RandomValue",
			Value: rand.Float64(),
		}
	}())
	r.PollCounter.Value += 1

	return err
}
