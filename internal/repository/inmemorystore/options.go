package inmemorystore

import (
	"time"

	"github.com/OmAsana/yapraktikum/internal/logging"
)

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

func WithLogger(logger *logging.Logger) Options {
	return func(server *InMemoryStore) {
		server.log = logger
	}
}

func WithRestore(restore bool) Options {
	return func(server *InMemoryStore) {
		server.restore = restore
	}
}
