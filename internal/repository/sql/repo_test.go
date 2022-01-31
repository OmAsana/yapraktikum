package sql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/OmAsana/yapraktikum/internal/metrics"
)

func TestNewRepository(t *testing.T) {
	repo, err := NewRepository("postgres://praktikum:12345@10.173.138.47:5432/praktikum?sslmode=disable")
	require.NoError(t, err)

	repo.Ping()

	err = repo.StoreCounter(metrics.Counter{
		Name:  "counter",
		Value: 123123,
	})
	require.NoError(t, err)

	err = repo.StoreGauge(metrics.Gauge{
		Name:  "gauge",
		Value: 1.0123,
	})
	require.NoError(t, err)

	c, err := repo.RetrieveCounter("counter")
	require.NoError(t, err)
	fmt.Println(c)

	g, err := repo.RetrieveGauge("gauge")
	require.NoError(t, err)
	fmt.Println(g)

	gauge, counters, err := repo.ListStoredMetrics()
	require.NoError(t, err)
	fmt.Println(gauge, counters)

}
