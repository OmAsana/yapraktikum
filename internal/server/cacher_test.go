package server

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/OmAsana/yapraktikum/internal/handlers"
	"github.com/OmAsana/yapraktikum/internal/pkg"
)

func TestNewCacher(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "cacher_test_file")
	//defer os.Remove(file.Name())
	assert.NoError(t, err)
	fmt.Println(file.Name())

	repo := NewRepositoryMock()
	cacher, err := NewCacher(file.Name(), repo)
	assert.NoError(t, err)

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

	err = cacher.WriteMetric(&data)
	assert.NoError(t, err)

	m, err := cacher.ReadMetricsFromCache()
	assert.NoError(t, err)

	fmt.Println(m)

}
