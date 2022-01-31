package metrics

import (
	"fmt"
	"reflect"
	"runtime"
)

var memoryStats = []string{
	"Alloc",
	"TotalAlloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCCPUFraction",
	"GCSys",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
}

func CollectRuntimeMetrics() ([]Gauge, error) {
	memoryStats, err := memStats()
	if err != nil {
		return nil, err
	}
	return memoryStats, nil
}

func memStats() ([]Gauge, error) {
	mStats := new(runtime.MemStats)
	runtime.ReadMemStats(mStats)

	var metricsSlice []Gauge
	for _, v := range memoryStats {
		gauge, err := reflectMemoryStats(mStats, v)
		if err != nil {
			return nil, err
		}

		metricsSlice = append(metricsSlice, gauge)
	}

	return metricsSlice, nil
}

func reflectMemoryStats(stats *runtime.MemStats, fieldName string) (Gauge, error) {
	val := reflect.ValueOf(stats).Elem()
	for i := 0; i < val.NumField(); i++ {
		if fieldName == val.Type().Field(i).Name {
			switch v := val.Field(i).Interface().(type) {
			case float64:
				return Gauge{
					Name:  fieldName,
					Value: v,
				}, nil
			case uint32:
				return Gauge{
					Name:  fieldName,
					Value: float64(v),
				}, nil
			case uint64:
				return Gauge{
					Name:  fieldName,
					Value: float64(v),
				}, nil
			}

		}
	}
	return Gauge{}, fmt.Errorf("could not find memory stats %s", fieldName)
}
