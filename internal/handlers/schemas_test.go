package handlers

import (
	"encoding/json"
	"testing"

	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/OmAsana/yapraktikum/internal/pkg"
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

func TestMetrics_ComputeHash(t *testing.T) {
	key := "blabla"
	m := Metrics{
		ID:    "gauge",
		MType: "gauge",
		Delta: nil,
		Value: pkg.PointerFloat(12),
		Hash:  "",
	}
	m2 := Metrics{}
	err := copier.Copy(&m2, m)
	assert.NoError(t, err)

	h1, err := m2.ComputeHash(key)
	require.NoError(t, err)
	h2, err := m.ComputeHash(key)
	require.NoError(t, err)

	require.Equal(t, h1, h2)

}
