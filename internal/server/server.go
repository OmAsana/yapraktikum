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
}

type ServerOpts func(server *MetricsServer)

func WithStoreFile(file string) ServerOpts {
	return func(server *MetricsServer) {
		server.storeFile = file
	}
}

func WithStoreInterval(interval time.Duration) ServerOpts {
	return func(server *MetricsServer) {
		server.storeInterval = interval
	}
}

func WithRestore(restore bool) ServerOpts {
	return func(server *MetricsServer) {
		server.restore = restore
	}
}

func NewMetricsServer(db MetricsRepository, opts ...ServerOpts) (*MetricsServer, error) {
	srv := &MetricsServer{
		db:            db,
		Mux:           chi.NewMux(),
		storeInterval: 5 * time.Second,
		storeFile:     "/tmp/devops-metrics-db.json",
		restore:       true,
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

func (receiver MetricsServer) Value() http.HandlerFunc {
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
			c, err := receiver.db.RetrieveCounter(m.ID)
			if err != nil {
				fmt.Printf("Not found metric %+v", m)
				http.Error(writer, err.Error(), http.StatusNotFound)
				return
			}

			m.Delta = &c.Value

		case "gauge":
			g, err := receiver.db.RetrieveGauge(m.ID)
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

func (receiver MetricsServer) Update() http.HandlerFunc {
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
		receiver.saveMetric(writer, m)
	}
}

func (receiver MetricsServer) saveMetric(writer http.ResponseWriter, m handlers.Metrics) {

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

		err := receiver.db.StoreCounter(c)
		if err != nil {
			http.Error(writer, "internal error", http.StatusInternalServerError)
			return
		}
		receiver.writeMetricToFile(&m)
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

		err := receiver.db.StoreGauge(g)
		if err != nil {
			http.Error(writer, "internal error", http.StatusInternalServerError)
			return
		}
		receiver.writeMetricToFile(&m)
		writer.WriteHeader(http.StatusOK)
		return
	default:
		http.Error(writer, "wrong metric type", http.StatusBadRequest)
	}
	http.Error(writer, "not implemented", http.StatusNotImplemented)
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
		metric := handlers.Metrics{
			ID:    metricName,
			MType: "gauge",
			Delta: nil,
			Value: &val,
		}
		receiver.saveMetric(writer, metric)
	}
}

func (receiver MetricsServer) UpdateCounters() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		metricName := chi.URLParam(request, "counterName")
		value := chi.URLParam(request, "counterValue")

		val, err := validateCounter(value)
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
		receiver.saveMetric(writer, metric)
	}

}

func validateCounter(value string) (int64, error) {
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return val, fmt.Errorf("value is not int")

	}

	if val < 0 {
		return val, fmt.Errorf("counter can not be negative")
	}
	return val, err
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

func (receiver MetricsServer) restoreData() error {
	reader, err := NewCacherReader(receiver.storeFile)
	if err != nil {
		return err
	}
	for {
		m, err := reader.ReadMetricsFromCache()
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF {
			break
		}

		defer func() {
			//if err := recover(); err != nil {
			//	fmt.Printf("%+v\n", m)
			//}
		}()
		switch m.MType {
		case "counter":

			c := metrics.Counter{
				Name:  m.ID,
				Value: *m.Delta,
			}

			err := receiver.db.StoreCounter(c)
			if err != nil {
				return err
			}
		case "gauge":

			g := metrics.Gauge{
				Name:  m.ID,
				Value: *m.Value,
			}
			err := receiver.db.StoreGauge(g)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (receiver MetricsServer) writeMetricToFile(m *handlers.Metrics) {
	if receiver.storeInterval > 0 {
		return
	}

	err := receiver.cacherWriter.WriteMetric(m)
	if err != nil {
		fmt.Println(err)
	}
}

func (receiver MetricsServer) periodicDataWriter() {
	if receiver.storeInterval > 0 {
		go func() {
			ticker := time.NewTicker(receiver.storeInterval)
			for range ticker.C {
				receiver.flushToDisk()
			}
		}()
	}
}

func (receiver MetricsServer) flushToDisk() {
	gauges, couters, err := receiver.db.ListStoredMetrics()
	if err != nil {
		fmt.Println(err)
	}
	for _, g := range gauges {
		m := &handlers.Metrics{
			ID:    g.Name,
			MType: "gauge",
			Delta: nil,
			Value: &g.Value,
		}

		err := receiver.cacherWriter.WriteMetric(m)
		if err != nil {
			fmt.Println(err)
		}
	}

	for _, c := range couters {
		m := &handlers.Metrics{
			ID:    c.Name,
			MType: "counter",
			Delta: &c.Value,
			Value: nil,
		}

		err := receiver.cacherWriter.WriteMetric(m)
		if err != nil {
			fmt.Println(err)
		}

	}
}

func (receiver MetricsServer) FlushToDisk() {
	receiver.flushToDisk()
}
