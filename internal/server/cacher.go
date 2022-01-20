package server

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/OmAsana/yapraktikum/internal/handlers"
)

type Cacher interface {
	WriteMetric(m *[]handlers.Metrics) error
	Close() error
}

var _ Cacher = (*cacher)(nil)

type cacher struct {
	file          *os.File
	encoder       *json.Encoder
	decoder       *json.Decoder
	db            MetricsRepository
	requestStream chan writeMetricsRequest
	doOnce        *sync.Once
}

func NewCacher(fileName string, db MetricsRepository) (*cacher, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}

	incomingMetricChan := make(chan writeMetricsRequest)

	return &cacher{
		file:          file,
		encoder:       json.NewEncoder(file),
		decoder:       json.NewDecoder(file),
		db:            db,
		requestStream: incomingMetricChan,
		doOnce:        &sync.Once{},
	}, nil
}

func (c *cacher) trucateFile() error {
	err := os.Truncate(c.file.Name(), 0)
	return err

}

func (c *cacher) WriteMetric(metrics *[]handlers.Metrics) error {
	errChan := make(chan error)
	request := writeMetricsRequest{
		metrics: metrics,
		err:     errChan,
	}

	c.writeMetricChan() <- request

	return <-errChan
}

type writeMetricsRequest struct {
	metrics *[]handlers.Metrics
	err     chan<- error
}

func (c *cacher) writeMetricChan() chan writeMetricsRequest {

	c.doOnce.Do(func() {
		go func() {
			defer c.file.Close()
			for {
				select {
				case metrics, open := <-c.requestStream:
					if !open {
						return
					}
					if err := c.trucateFile(); err != nil {
						metrics.err <- err
						return
					}

					for _, m := range *metrics.metrics {
						err := c.encoder.Encode(&m)
						if err != nil {
							metrics.err <- err
							return
						}
						fmt.Println("blah")
					}

					close(metrics.err)
				}
			}
		}()
	})
	return c.requestStream

}

func (c *cacher) ReadMetricsFromCache() ([]handlers.Metrics, error) {
	metrics := []handlers.Metrics{}

	_, err := c.file.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	for {
		var m handlers.Metrics
		err := c.decoder.Decode(&m)
		if err == io.EOF {
			return metrics, nil
		}
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

}

func (c *cacher) Close() error {
	return c.file.Close()

}

var _ Cacher = (*noopCacher)(nil)

type noopCacher struct {
}

func NewNoopCacher() *noopCacher {
	return &noopCacher{}
}

func (n *noopCacher) WriteMetric(_ *[]handlers.Metrics) error {
	return nil
}

func (n *noopCacher) Close() error {
	return nil
}
