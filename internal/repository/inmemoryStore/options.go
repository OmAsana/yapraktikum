package inmemoryStore

import "time"

type Options func(server *InMemoryStore)

func WithStoreFile(file string) Options {
	return func(server *InMemoryStore) {
		server.storeFile = file
	}
}

func WithStoreInterval(interval time.Duration) Options {
	return func(server *InMemoryStore) {
		server.storeInterval = interval
	}
}

func WithRestore(restore bool) Options {
	return func(server *InMemoryStore) {
		server.restore = restore
	}
}
