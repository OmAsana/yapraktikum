package server

type Options func(server *MetricsServer)

func WithHashKey(key string) Options {
	return func(server *MetricsServer) {
		server.hashKey = key
	}
}
