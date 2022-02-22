package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_CollectRuntimeMetrics(t *testing.T) {
	gauges, err := CollectRuntimeMetrics(context.Background())
	require.NoError(t, err)

	wantStats := memoryStats
	wantStats = append(wantStats, "FreeMemory", "TotalMemory")
	for _, statName := range wantStats {
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
