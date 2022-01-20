package metrics

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGauge(t *testing.T) {
	t.Run("test json marshaling", func(t *testing.T) {

		gauge := Gauge{
			Name:  "Blah",
			Value: 4,
		}

		d, err := json.Marshal(&gauge)
		assert.NoError(t, err)
		assert.Equal(t, string(d), `{"mType":"gauge","name":"Blah","value":4}`)

		newGauge := Gauge{}
		err = json.Unmarshal(d, &newGauge)
		assert.NoError(t, err)
		assert.Equal(t, gauge, newGauge)

	})

}
