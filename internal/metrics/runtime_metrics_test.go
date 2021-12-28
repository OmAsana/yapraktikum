package metrics

import (
	"testing"
)

func Test_CollectRuntimeMetrics(t *testing.T) {
	gauges := CollectRuntimeMetrics()
	for _, statName := range memoryStats {
		found := false
		for _, gauge := range gauges {
			if gauge.Name == statName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("metric not found: %s", statName)
		}
	}
}
