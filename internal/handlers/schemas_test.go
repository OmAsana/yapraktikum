package handlers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics(t *testing.T) {
	t.Run("Test gauge encode/decode", func(t *testing.T) {
		mockJSONValue := `{
"id": "Blah",
"type": "gauge",
"value": 1.02
}`
		wantVal := 1.02
		var want = Metrics{
			ID:    "Blah",
			MType: "gauge",
			Value: &wantVal,
		}

		t.Run("decode", func(t *testing.T) {
			var m Metrics
			err := json.Unmarshal([]byte(mockJSONValue), &m)
			require.NoError(t, err)
			assert.Equal(t, want, m)
		})
		t.Run("encode", func(t *testing.T) {
			out, err := json.MarshalIndent(want, "", "")
			require.NoError(t, err)
			assert.Equal(t, mockJSONValue, string(out))

		})
	})

	t.Run("Test counter encode/decode", func(t *testing.T) {
		mockJSONValue := `{
"id": "Blah",
"type": "counter",
"delta": 22
}`
		wantVal := int64(22)
		var want = Metrics{
			ID:    "Blah",
			MType: "counter",
			Delta: &wantVal,
		}
		t.Run("decode", func(t *testing.T) {
			var m Metrics
			err := json.Unmarshal([]byte(mockJSONValue), &m)
			require.NoError(t, err)
			assert.Equal(t, want, m)
		})
		t.Run("encode", func(t *testing.T) {
			out, err := json.MarshalIndent(want, "", "")
			require.NoError(t, err)
			assert.Equal(t, mockJSONValue, string(out))
		})
	})

	t.Run("Test wrong values", func(t *testing.T) {
		mockJSONValue := `{
"some": "blah",
"random": "blah",
"json": "blah"
}`
		t.Run("decode", func(t *testing.T) {
			var m Metrics
			err := json.Unmarshal([]byte(mockJSONValue), &m)
			require.Error(t, err)
		})

	})

}