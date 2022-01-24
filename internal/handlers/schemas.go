package handlers

import (
	"encoding/json"
	"fmt"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (m *Metrics) UnmarshalJSON(bytes []byte) error {
	//TODO implement me

	type MetricsAlias Metrics
	aliasValue := &struct {
		*MetricsAlias
	}{
		MetricsAlias: (*MetricsAlias)(m),
	}

	if err := json.Unmarshal(bytes, &aliasValue); err != nil {
		return err
	}

	if m.ID == "" || m.MType == "" {
		return fmt.Errorf("missing required fields")
	}
	return nil
}
