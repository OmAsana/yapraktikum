package inmemory_store

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/OmAsana/yapraktikum/internal/handlers"
	"github.com/OmAsana/yapraktikum/internal/metrics"
	"github.com/OmAsana/yapraktikum/internal/repository"
)

var _ repository.MetricsRepository = (*InMemoryStore)(nil)

type InMemoryStore struct {
	sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64

	cacherWriter Cacher
	cacherReader *CacherReader

	storeInterval time.Duration
	storeFile     string
	restore       bool

	storeSignal chan struct{}
}

func (r *InMemoryStore) WriteBulkGauges(gauges []metrics.Gauge) error {
	for _, g := range gauges {
		if err := r.StoreGauge(g); err != nil {
			return err
		}
	}
	return nil
}

func (r *InMemoryStore) WriteBulkCounters(counters []metrics.Counter) error {
	for _, c := range counters {
		if err := r.StoreCounter(c); err != nil {
			return err
		}
	}
	return nil
}

func (r *InMemoryStore) Ping() bool {
	return false
}

func NewDefaultInMemoryRepo() *InMemoryStore {
	repo := &InMemoryStore{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}

	return repo
}

func NewInMemoryRepo(opts ...Options) (*InMemoryStore, error) {
	repo := &InMemoryStore{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),

		storeInterval: 0 * time.Second,
		storeFile:     "",
		restore:       false,
		storeSignal:   make(chan struct{}),
	}
	for _, opt := range opts {
		opt(repo)
	}

	if repo.storeFile != "" {
		if repo.restore {
			if err := repo.restoreData(); err != nil {
				return nil, err
			}
		}

		cacheWriter, err := NewCacherWriter(repo.storeFile)
		if err != nil {
			return nil, err
		}
		repo.cacherWriter = cacheWriter
	} else {
		repo.cacherWriter = NewNoopCacher()
	}

	go repo.flushToDiskRoutine()

	return repo, nil
}

func (r *InMemoryStore) flushToDiskRoutine() {
	if r.storeInterval > 0 {
		go func() {
			ticker := time.NewTicker(r.storeInterval)
			for {
				select {
				case <-ticker.C:
					r.flushToDisk()
				case <-r.storeSignal:
					continue
				}
			}
		}()
	} else {
		go func() {
			for range r.storeSignal {
				r.flushToDisk()
			}
		}()
	}

}

func (r *InMemoryStore) RetrieveCounter(name string) (metrics.Counter, repository.RepositoryError) {
	r.RLock()
	defer r.RUnlock()
	if v, ok := r.counters[name]; ok {
		return metrics.Counter{
			Name:  name,
			Value: v,
		}, nil
	}
	return metrics.Counter{}, repository.ErrorCounterNotFound
}

func (r *InMemoryStore) RetrieveGauge(name string) (metrics.Gauge, repository.RepositoryError) {
	r.RLock()
	defer r.RUnlock()
	if v, ok := r.gauges[name]; ok {
		return metrics.Gauge{
			Name:  name,
			Value: v,
		}, nil
	}
	return metrics.Gauge{}, repository.ErrorGaugeNotFound
}

func (r *InMemoryStore) StoreCounter(counter metrics.Counter) repository.RepositoryError {
	r.Lock()
	defer r.Unlock()
	err := counter.IsValid()
	if err != nil {
		return repository.ErrorCounterIsNoValid
	}

	_, ok := r.counters[counter.Name]
	if ok {
		r.counters[counter.Name] += counter.Value
		return nil
	}
	r.counters[counter.Name] = counter.Value

	return nil
}

func (r *InMemoryStore) StoreGauge(gauge metrics.Gauge) repository.RepositoryError {
	r.Lock()
	defer r.Unlock()
	r.gauges[gauge.Name] = gauge.Value
	return nil
}

func (r *InMemoryStore) ListStoredMetrics() ([]metrics.Gauge, []metrics.Counter, repository.RepositoryError) {
	var gauges []metrics.Gauge
	var couter []metrics.Counter

	r.RLock()
	defer r.RUnlock()
	for k, v := range r.gauges {
		gauges = append(gauges, metrics.Gauge{
			Name:  k,
			Value: v,
		})
	}

	for k, v := range r.counters {
		couter = append(couter, metrics.Counter{
			Name:  k,
			Value: v,
		})
	}

	return gauges, couter, nil
}

func (r *InMemoryStore) restoreData() error {
	reader, err := NewCacherReader(r.storeFile)
	if err != nil {
		return err
	}
	metricsFromDisk, err := reader.ReadMetricsFromCache()
	if err != nil && err != io.EOF {
		return err
	}
	if err == io.EOF {
		return nil
	}
	for _, m := range metricsFromDisk {
		switch m.MType {
		case "counter":

			c := metrics.Counter{
				Name:  m.ID,
				Value: *m.Delta,
			}

			err := r.StoreCounter(c)
			if err != nil {
				return err
			}
		case "gauge":

			g := metrics.Gauge{
				Name:  m.ID,
				Value: *m.Value,
			}
			err := r.StoreGauge(g)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (r *InMemoryStore) flushToDisk() {
	gauges, couters, err := r.ListStoredMetrics()
	if err != nil {
		fmt.Println(err)
	}
	flushMetrics := []handlers.Metrics{}
	for _, g := range gauges {
		m := metrics.GaugeToHandlerScheme(g)
		flushMetrics = append(flushMetrics, m)

	}

	for _, c := range couters {
		m := metrics.CounterToHandlerScheme(c)
		flushMetrics = append(flushMetrics, m)

	}

	err = r.cacherWriter.WriteMultipleMetrics(&flushMetrics)
	if err != nil {
		fmt.Println(err)
	}
}
