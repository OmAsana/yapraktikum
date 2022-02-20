package metrics

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/sync/errgroup"
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

type systemStatCollector func() ([]Gauge, error)

func CollectRuntimeMetrics() ([]Gauge, error) {
	g, _ := errgroup.WithContext(context.Background())

	collectors := []systemStatCollector{runtimeMemStats, totalAndFreeMem}
	var result []Gauge
	writeMu := sync.Mutex{}

	for _, collector := range collectors {
		collector := collector
		g.Go(func() error {
			stats, err := collector()
			if err == nil {
				writeMu.Lock()
				result = append(result, stats...)
				writeMu.Unlock()
			}
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

func totalAndFreeMem() ([]Gauge, error) {
	memStats, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	return []Gauge{
			{
				Name:  "FreeMemory",
				Value: float64(memStats.Free)},
			{
				Name:  "TotalMemory",
				Value: float64(memStats.Total)}},
		nil
}

func runtimeMemStats() ([]Gauge, error) {
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
