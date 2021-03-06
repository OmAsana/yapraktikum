package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/OmAsana/yapraktikum/internal/encrypt"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func (m *Metrics) UnmarshalJSON(bytes []byte) error {
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

func (m *Metrics) HashMetric(key string) error {
	encrypted, err := m.ComputeHash(key)
	if err != nil {
		return err
	}

	m.Hash = encrypted
	return nil

}

func (m *Metrics) ComputeHash(key string) (string, error) {
	var encrypted string

	if m.Delta != nil {
		encrypted = encrypt.EncryptSHA256(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta), key)
	}

	if m.Value != nil {
		encrypted = encrypt.EncryptSHA256(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value), key)
	}

	if encrypted == "" {
		return "", fmt.Errorf("invalid metric")
	}
	return encrypted, nil
}
