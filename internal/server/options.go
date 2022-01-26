package server

import "time"

type Options func(server *MetricsServer)

func WithStoreFile(file string) Options {
	return func(server *MetricsServer) {
		server.storeFile = file
	}
}

func WithStoreInterval(interval time.Duration) Options {
	return func(server *MetricsServer) {
		server.storeInterval = interval
	}
}

func WithRestore(restore bool) Options {
	return func(server *MetricsServer) {
		server.restore = restore
	}
}

func WithHashKey(key string) Options {
	return func(server *MetricsServer) {
		server.hashKey = key
	}
}
