package server

import "time"

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
