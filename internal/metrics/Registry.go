package metrics

import (
	"math/rand"
	"sync"
)

type Registry struct {
	sync.RWMutex
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
	r.Lock()
	defer r.Unlock()

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

func (r *Registry) Export() ([]Gauge, []Counter) {
	r.RLock()
	defer r.RUnlock()

	counter := r.Counters
	counter = append(counter, r.PollCounter)
	return r.Gauges, counter
}
