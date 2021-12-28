package server

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/OmAsana/yapraktikum/internal/metrics"
	"github.com/OmAsana/yapraktikum/internal/pkg"
)

type MetricsServer struct {
	*chi.Mux
	db MetricsRepository
}

func NewMetricsServer(db MetricsRepository) *MetricsServer {
	srv := &MetricsServer{
		db:  db,
		Mux: chi.NewMux()}

	srv.Use(middleware.RequestID)
	srv.Use(middleware.RealIP)
	srv.Use(middleware.Logger)
	srv.Use(middleware.Recoverer)

	srv.Get("/", srv.ReturnCurrentMetrics())
	srv.Get("/value/{metricType}/{metricName}", srv.GetMetric())

	srv.Route("/update", func(r chi.Router) {
		r.Route("/counter/", func(r chi.Router) {
			r.Post("/{counterName}/{counterValue}", srv.UpdateCounters())
		})
		r.Route("/gauge/", func(r chi.Router) {
			r.Post("/{gaugeName}/{gaugeValue}", srv.UpdateGauge())
		})
		r.Post("/*", func(writer http.ResponseWriter, request *http.Request) {
			http.Error(writer, "not implemented", http.StatusNotImplemented)
		})

	})

	return srv
}

func (receiver MetricsServer) GetMetric() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		metricType := chi.URLParam(request, "metricType")
		metricName := chi.URLParam(request, "metricName")
		switch metricType {
		case "counter":
			receiver.writeCounter(writer, metricName)
		case "gauge":
			receiver.writeGauge(writer, metricName)
		default:
			http.Error(writer, "", http.StatusNotFound)
		}
	}
}

func (receiver MetricsServer) writeGauge(writer http.ResponseWriter, gaugeName string) {
	val, err := receiver.db.RetrieveGauge(gaugeName)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusNotFound)
		return
	}
	_, err = io.WriteString(writer, strconv.FormatFloat(val.Value, 'g', -1, 64))
	if err != nil {
		http.Error(writer, "internal error", http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func (receiver MetricsServer) writeCounter(writer http.ResponseWriter, counterName string) {
	val, err := receiver.db.RetrieveCounter(counterName)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusNotFound)
		return
	}
	_, err = io.WriteString(writer, strconv.FormatInt(val.Value, 10))
	if err != nil {
		http.Error(writer, "internal error", http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func (receiver MetricsServer) UpdateGauge() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		metricName := chi.URLParam(request, "gaugeName")
		value := chi.URLParam(request, "gaugeValue")

		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(writer, "value is not float", http.StatusBadRequest)
			return
		}

		if !pkg.FloatIsNumber(val) {
			http.Error(writer, "float must be a number", http.StatusBadRequest)
			return
		}

		err = receiver.db.StoreGauge(metrics.Gauge{
			Name:  metricName,
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
		metricName := chi.URLParam(request, "counterName")
		value := chi.URLParam(request, "counterValue")

		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(writer, "value is not int", http.StatusBadRequest)
			return

		}

		if val < 0 {
			http.Error(writer, "counter can not be negative", http.StatusBadRequest)
			return
		}

		err = receiver.db.StoreCounter(metrics.Counter{
			Name:  metricName,
			Value: val,
		})

		if err != nil {
			http.Error(writer, "internal error", http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}
}

func (receiver MetricsServer) ReturnCurrentMetrics() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var sb strings.Builder

		gauges, counters, err := receiver.db.ListStoredMetrics()
		if err != nil {
			http.Error(writer, "internal error", http.StatusInternalServerError)
		}
		for _, g := range gauges {
			sb.WriteString(fmt.Sprintf("%s\t\t%f\n", g.Name, g.Value))
		}
		for _, c := range counters {
			sb.WriteString(fmt.Sprintf("%s\t\t%d\n", c.Name, c.Value))
		}
		_, err = io.WriteString(writer, sb.String())
		if err != nil {
			http.Error(writer, "internal error", http.StatusInternalServerError)
		}

		sb.Reset()
		writer.WriteHeader(http.StatusOK)
	}
}
