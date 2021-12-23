package server

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/OmAsana/yapraktikum/internal/metrics"
)

type MetricsServer struct {
	db MetricsRepository
}

func NewMetricsServer(db MetricsRepository) *MetricsServer {
	return &MetricsServer{db: db}
}

func (receiver MetricsServer) UpdateGauge() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "method not supported", http.StatusBadRequest)
			return
		}
		path := strings.Split(strings.TrimPrefix(request.URL.Path, "/"), "/")
		if len(path) != 4 {
			http.Error(writer, "invalid parameters", http.StatusBadRequest)
			return
		}

		if path[1] != "gauge" {
			http.Error(writer, "invalid parameters", http.StatusBadRequest)
			return
		}

		val, err := strconv.ParseFloat(path[3], 64)
		if err != nil {
			http.Error(writer, "value is not float", http.StatusBadRequest)
			return
		}

		if !floatIsNumber(val) {
			http.Error(writer, "float must be a number", http.StatusBadRequest)
			return
		}

		err = receiver.db.StoreGauge(metrics.Gauge{
			Name:  path[2],
			Value: val,
		})

		if err != nil {
			http.Error(writer, "internal error", http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}

}

func (receiver MetricsServer) UpdateCounters() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		if request.Method != http.MethodPost {
			http.Error(writer, "method not supported", http.StatusBadRequest)
			return
		}

		path := strings.Split(strings.TrimPrefix(request.URL.Path, "/"), "/")
		if len(path) != 4 {
			http.Error(writer, "invalid parameters", http.StatusBadRequest)
			return
		}

		if path[1] != "counter" {
			http.Error(writer, "invalid parameters", http.StatusBadRequest)
			return
		}

		val, err := strconv.ParseInt(path[3], 10, 64)
		if err != nil {
			http.Error(writer, "value is not int", http.StatusBadRequest)
			return

		}

		if val < 0 {
			http.Error(writer, "counter can not be negative", http.StatusBadRequest)
			return
		}

		err = receiver.db.StoreCounter(metrics.Counter{
			Name:  path[2],
			Value: val,
		})

		if err != nil {
			http.Error(writer, "internal error", http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}
}

//floatIsNumber check that f is not an inf of Nan
func floatIsNumber(f float64) bool {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return false
	}
	return true
}
