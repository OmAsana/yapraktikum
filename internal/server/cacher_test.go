package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/OmAsana/yapraktikum/internal/handlers"
	"github.com/OmAsana/yapraktikum/internal/pkg"
)

func TestNewCacher(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "cacher_test_file")
	defer os.Remove(file.Name())
	assert.NoError(t, err)
	fmt.Println(file.Name())

	cacher, err := NewCacherWriter(file.Name())
	assert.NoError(t, err)
	defer cacher.Close()

	rand.Seed(time.Now().UnixNano())

	data := []handlers.Metrics{
		{
			ID:    "gauge2",
			MType: "gauge",
			Delta: pkg.PointerInt(int64(rand.Int())),
			Value: nil,
		},
		{
			ID:    "couter1",
			MType: "counter",
			Delta: nil,
			Value: pkg.PointerFloat(rand.Float64()),
		},
	}

	for _, m := range data {
		err = cacher.WriteMetric(&m)
		assert.NoError(t, err)
	}

	reader, err := NewCacherReader(file.Name())
	assert.NoError(t, err)
	defer reader.Close()

	metrics := []handlers.Metrics{}

	for {
		m, err := reader.ReadMetricsFromCache()
		if err != nil && err != io.EOF {
			assert.NoError(t, err)
		}
		if err == io.EOF {
			break
		}
		metrics = append(metrics, m)
	}

	for k, v := range data {
		assert.ObjectsAreEqualValues(v, metrics[k])
	}

}
