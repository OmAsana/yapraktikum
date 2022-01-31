package sql

import "github.com/OmAsana/yapraktikum/internal/logging"

type Option func(*Repository) error

func WithLogger(l *logging.Logger) Option {
	return func(repository *Repository) error {
		repository.log = l
		return nil
	}
}
