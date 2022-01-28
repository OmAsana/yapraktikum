package mock

import "time"

type Options func(server *RepositoryMock)

func WithStoreFile(file string) Options {
	return func(server *RepositoryMock) {
		server.storeFile = file
	}
}

func WithStoreInterval(interval time.Duration) Options {
	return func(server *RepositoryMock) {
		server.storeInterval = interval
	}
}

func WithRestore(restore bool) Options {
	return func(server *RepositoryMock) {
		server.restore = restore
	}
}
