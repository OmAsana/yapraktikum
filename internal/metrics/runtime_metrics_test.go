package metrics

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

func TestCollectRuntimeMetrics(t *testing.T) {
}

func Test_memStats(t *testing.T) {
	mStats := new(runtime.MemStats)
	runtime.ReadMemStats(mStats)
	val := reflect.ValueOf(mStats).Elem()
	for i := 0; i < val.NumField(); i++ {
		switch v := val.Field(i).Interface().(type) {
		case uint64:
			fmt.Println(v)
		}
		//fmt.Println(val.Type().Field(i).Name)
	}
}

func Test_CollectRuntimeMetrics(t *testing.T) {
	CollectRuntimeMetrics()
}
