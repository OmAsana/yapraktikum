package inmemorystore

import (
	"encoding/json"
	"os"

	"github.com/OmAsana/yapraktikum/internal/handlers"
)

type CacheWriter interface {
	WriteSingleMetric(m *handlers.Metrics) error
	WriteMultipleMetrics(m *[]handlers.Metrics) error
	Close() error
}

var _ CacheWriter = (*cacheWriter)(nil)

type cacheWriter struct {
	file    *os.File
	encoder *json.Encoder
}

func NewCacherWriter(fileName string) (*cacheWriter, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &cacheWriter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (c *cacheWriter) WriteSingleMetric(metrics *handlers.Metrics) error {
	return c.encoder.Encode(&metrics)
}

func (c cacheWriter) WriteMultipleMetrics(metrics *[]handlers.Metrics) error {
	os.Truncate(c.file.Name(), 0)
	c.file.Seek(0, 0)
	return c.encoder.Encode(metrics)
}

func (c *cacheWriter) Close() error {
	return c.file.Close()
}

type CacheReader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewCacherReader(fileName string) (*CacheReader, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &CacheReader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *CacheReader) Close() error {
	return c.file.Close()

}

func (c *CacheReader) ReadMetricsFromCache() ([]handlers.Metrics, error) {
	var m []handlers.Metrics
	err := c.decoder.Decode(&m)
	return m, err
}

func (c *CacheReader) TruncateFile() error {
	return os.Truncate(c.file.Name(), 0)
}

var _ CacheWriter = (*noopCacher)(nil)

type noopCacher struct {
}

func NewNoopCacher() *noopCacher {
	return &noopCacher{}
}

func (n *noopCacher) WriteMultipleMetrics(_ *[]handlers.Metrics) error {
	return nil
}

func (n *noopCacher) WriteSingleMetric(_ *handlers.Metrics) error {
	return nil
}

func (n *noopCacher) Close() error {
	return nil
}
