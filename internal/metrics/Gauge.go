package metrics

import (
	"encoding/json"
	"fmt"

	"github.com/OmAsana/yapraktikum/internal/handlers"
)

type Gauge struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

func (c *Gauge) MarshalJSON() ([]byte, error) {
	type GaugeAlias Gauge
	return json.Marshal(&struct {
		MetricType string `json:"mType"`
		*GaugeAlias
	}{
		MetricType: "gauge",
		GaugeAlias: (*GaugeAlias)(c),
	})
}

func (c Gauge) String() string {
	return fmt.Sprintf("<Gauge: Name: %s, Value: %f>", c.Name, c.Value)
}

func GaugeToHandlerScheme(g Gauge) handlers.Metrics {
	return handlers.Metrics{
		ID:    g.Name,
		MType: "gauge",
		Delta: nil,
		Value: &g.Value,
	}
}

func GaugeFromHandler(g handlers.Metrics) Gauge {
	return Gauge{
		Name:  g.ID,
		Value: *g.Value,
	}
}
