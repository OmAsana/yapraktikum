package metrics

import (
	"encoding/json"
	"fmt"
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
