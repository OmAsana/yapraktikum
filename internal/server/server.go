package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/OmAsana/yapraktikum/internal/handlers"
	"github.com/OmAsana/yapraktikum/internal/metrics"
	"github.com/OmAsana/yapraktikum/internal/pkg"
)

type MetricsServer struct {
	*chi.Mux
	db            MetricsRepository
	storeInterval time.Duration
	storeFile     string
	restore       bool
	cacherReader  *cacherReader
	cacherWriter  Cacher
	hashKey       string
}

func NewMetricsServer(db MetricsRepository, opts ...Options) (*MetricsServer, error) {
	srv := &MetricsServer{
		db:            db,
		Mux:           chi.NewMux(),
		storeInterval: 0 * time.Second,
		storeFile:     "0",
		restore:       false,
	}

	for _, opt := range opts {
		opt(srv)
	}

	if srv.storeFile != "" {
		if srv.restore {
			srv.restoreData()
		}

		cacheWriter, err := NewCacherWriter(srv.storeFile)
		if err != nil {
			return nil, err
		}
		srv.cacherWriter = cacheWriter
	} else {
		srv.cacherWriter = NewNoopCacher()
	}

	srv.periodicDataWriter()
	setupRoutes(srv)

	return srv, nil
}

func setupRoutes(srv *MetricsServer) {
	srv.Use(middleware.RequestID)
	srv.Use(middleware.RealIP)
	srv.Use(middleware.Logger)
	srv.Use(middleware.Recoverer)
	srv.Use(compressorHandler)

	srv.Get("/", srv.ReturnCurrentMetrics())
	srv.Get("/value/{metricType}/{metricName}", srv.GetMetric())

	srv.Post("/value/", srv.Value())

	srv.Route("/update", func(r chi.Router) {
		r.Post("/", srv.Update())
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
}

func (ms MetricsServer) Value() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if !pkg.Contains(request.Header.Values("accept"), "application/json") {
			http.Error(writer, "not implemented", http.StatusNotImplemented)
		}

		var m handlers.Metrics
		err := json.NewDecoder(request.Body).Decode(&m)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		switch m.MType {
		case "counter":
			c, err := ms.db.RetrieveCounter(m.ID)
			if err != nil {
				fmt.Printf("Not found metric %+v", m)
				http.Error(writer, err.Error(), http.StatusNotFound)
				return
			}

			m.Delta = &c.Value

		case "gauge":
			g, err := ms.db.RetrieveGauge(m.ID)
			if err != nil {
				fmt.Printf("Not found metric %+v", m)
				http.Error(writer, err.Error(), http.StatusNotFound)
				return
			}

			m.Value = &g.Value

		default:
			http.NotFound(writer, request)
			return

		}

		out, err := json.Marshal(m)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Add("Content-Type", "application/json")
		writer.Write(out)
	}
}

func (ms MetricsServer) Update() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if !pkg.Contains(request.Header.Values("Content-Type"), "application/json") {
			http.Error(writer, "not implemented", http.StatusNotImplemented)
		}

		var m handlers.Metrics
		err := json.NewDecoder(request.Body).Decode(&m)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		ms.saveMetric(writer, m)
	}
}

func (ms MetricsServer) saveMetric(writer http.ResponseWriter, m handlers.Metrics) {

	switch m.MType {
	case "counter":
		if m.Delta == nil {
			http.Error(writer, "delta can not be nil", http.StatusBadRequest)
			return
		}

		c := metrics.Counter{
			Name:  m.ID,
			Value: *m.Delta,
		}

		err := ms.db.StoreCounter(c)
		if err != nil {
			http.Error(writer, "internal error", http.StatusInternalServerError)
			return
		}
		ms.flushToDisk()
		writer.WriteHeader(http.StatusOK)
		return
	case "gauge":
		if m.Value == nil {
			http.Error(writer, "value can not be nil", http.StatusBadRequest)
			return
		}

		g := metrics.Gauge{
			Name:  m.ID,
			Value: *m.Value,
		}

		err := ms.db.StoreGauge(g)
		if err != nil {
			http.Error(writer, "internal error", http.StatusInternalServerError)
			return
		}
		ms.flushToDisk()
		writer.WriteHeader(http.StatusOK)
		return
	default:
		http.Error(writer, "wrong metric type", http.StatusBadRequest)
	}
	http.Error(writer, "not implemented", http.StatusNotImplemented)
}

func (ms MetricsServer) GetMetric() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		metricType := chi.URLParam(request, "metricType")
		metricName := chi.URLParam(request, "metricName")
		switch metricType {
		case "counter":
			ms.writeCounter(writer, metricName)
		case "gauge":
			ms.writeGauge(writer, metricName)
		default:
			http.Error(writer, "", http.StatusNotFound)
		}
	}
}

func (ms MetricsServer) writeGauge(writer http.ResponseWriter, gaugeName string) {
	val, err := ms.db.RetrieveGauge(gaugeName)
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

func (ms MetricsServer) writeCounter(writer http.ResponseWriter, counterName string) {
	val, err := ms.db.RetrieveCounter(counterName)
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

func (ms MetricsServer) UpdateGauge() http.HandlerFunc {
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

		metric := handlers.Metrics{
			ID:    metricName,
			MType: "gauge",
			Delta: nil,
			Value: &val,
		}
		ms.saveMetric(writer, metric)
	}
}

func (ms MetricsServer) UpdateCounters() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		metricName := chi.URLParam(request, "counterName")
		value := chi.URLParam(request, "counterValue")

		val, err := pkg.ValidateCounter(value)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		metric := handlers.Metrics{
			ID:    metricName,
			MType: "counter",
			Delta: &val,
			Value: nil,
		}
		ms.saveMetric(writer, metric)
	}

}

func (ms MetricsServer) ReturnCurrentMetrics() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var sb strings.Builder

		gauges, counters, err := ms.db.ListStoredMetrics()
		if err != nil {
			http.Error(writer, "internal error", http.StatusInternalServerError)
		}
		for _, g := range gauges {
			sb.WriteString(fmt.Sprintf("%s\t\t%f\n", g.Name, g.Value))
		}
		for _, c := range counters {
			sb.WriteString(fmt.Sprintf("%s\t\t%d\n", c.Name, c.Value))
		}
		writer.Header().Set("Content-Type", "text/html")
		_, err = io.WriteString(writer, sb.String())
		if err != nil {
			http.Error(writer, "internal error", http.StatusInternalServerError)
		}

		sb.Reset()
		writer.WriteHeader(http.StatusOK)
	}
}

func (ms MetricsServer) restoreData() error {
	reader, err := NewCacherReader(ms.storeFile)
	if err != nil {
		return err
	}
	metricsFromDisk, err := reader.ReadMetricsFromCache()
	if err != nil && err != io.EOF {
		fmt.Println("restore data")
		fmt.Println(err)
		return err
	}
	if err == io.EOF {
		return nil
	}
	for _, m := range metricsFromDisk {
		switch m.MType {
		case "counter":

			c := metrics.Counter{
				Name:  m.ID,
				Value: *m.Delta,
			}

			err := ms.db.StoreCounter(c)
			if err != nil {
				return err
			}
		case "gauge":

			g := metrics.Gauge{
				Name:  m.ID,
				Value: *m.Value,
			}
			err := ms.db.StoreGauge(g)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ms MetricsServer) writeMetricToFile(m *handlers.Metrics) {
	if ms.storeInterval > 0 {
		return
	}
	ms.flushToDisk()
}

func (ms MetricsServer) periodicDataWriter() {
	if ms.storeInterval > 0 {
		go func() {
			ticker := time.NewTicker(ms.storeInterval)
			for range ticker.C {
				ms.flushToDisk()
			}
		}()
	}
}

func (ms MetricsServer) flushToDisk() {
	gauges, couters, err := ms.db.ListStoredMetrics()
	if err != nil {
		fmt.Println(err)
	}
	flushMetrics := []handlers.Metrics{}
	for _, g := range gauges {
		m := metrics.GaugeToHandlerScheme(g)
		flushMetrics = append(flushMetrics, m)

	}

	for _, c := range couters {
		m := metrics.CounterToHandlerScheme(c)
		flushMetrics = append(flushMetrics, m)

	}

	err = ms.cacherWriter.WriteMultipleMetrics(&flushMetrics)
	if err != nil {
		fmt.Println(err)
	}
}

func (ms MetricsServer) FlushToDisk() {
	ms.flushToDisk()
}
