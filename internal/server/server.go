package server

type MetricsServer struct {
	db MetricsRepository
}

func NewMetricsServer(db MetricsRepository) *MetricsServer {
	return &MetricsServer{db: db}
}
