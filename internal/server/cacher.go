package server

import (
	"encoding/json"
	"os"

	"github.com/OmAsana/yapraktikum/internal/handlers"
)

type Cacher interface {
	WriteSingleMetric(m *handlers.Metrics) error
	WriteMultipleMetrics(m *[]handlers.Metrics) error
	Close() error
}

var _ Cacher = (*cacherWriter)(nil)

type cacherWriter struct {
	file    *os.File
	encoder *json.Encoder
}

func NewCacherWriter(fileName string) (*cacherWriter, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &cacherWriter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (c *cacherWriter) WriteSingleMetric(metrics *handlers.Metrics) error {
	return c.encoder.Encode(&metrics)
}

func (c cacherWriter) WriteMultipleMetrics(metrics *[]handlers.Metrics) error {
	os.Truncate(c.file.Name(), 0)
	c.file.Seek(0, 0)
	return c.encoder.Encode(metrics)
}

func (c *cacherWriter) Close() error {
	return c.file.Close()
}

type cacherReader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewCacherReader(fileName string) (*cacherReader, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &cacherReader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *cacherReader) Close() error {
	return c.file.Close()

}

func (c *cacherReader) ReadMetricsFromCache() ([]handlers.Metrics, error) {
	var m []handlers.Metrics
	err := c.decoder.Decode(&m)
	return m, err
}

func (c *cacherReader) TrucateFile() error {
	return os.Truncate(c.file.Name(), 0)
}

var _ Cacher = (*noopCacher)(nil)

type noopCacher struct {
}

func (n *noopCacher) WriteMultipleMetrics(m *[]handlers.Metrics) error {
	return nil
}

func NewNoopCacher() *noopCacher {
	return &noopCacher{}
}

func (n *noopCacher) WriteSingleMetric(_ *handlers.Metrics) error {
	return nil
}

func (n *noopCacher) Close() error {
	return nil
}
