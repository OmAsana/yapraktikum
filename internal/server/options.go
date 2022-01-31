package server

import "github.com/OmAsana/yapraktikum/internal/logging"

type Options func(server *MetricsServer)

func WithHashKey(key string) Options {
	return func(server *MetricsServer) {
		server.hashKey = key
	}
}

func WithLogger(logger *logging.Logger) Options {
	return func(server *MetricsServer) {
		server.log = logger
	}
}
