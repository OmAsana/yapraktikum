package sql

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/OmAsana/yapraktikum/internal/metrics"
	"github.com/OmAsana/yapraktikum/internal/repository"
)

var _ repository.MetricsRepository = (*Repository)(nil)

type Repository struct {
	db *sql.DB
}

func NewRepository(dbn string) (*Repository, error) {
	db, err := sql.Open("pgx", dbn)
	if err != nil {
		return nil, err
	}
	return &Repository{
		db: db,
	}, nil
}

func (r *Repository) Ping() bool {
	err := r.db.Ping()
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func (r *Repository) StoreCounter(counter metrics.Counter) repository.RepositoryError {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) RetrieveCounter(name string) (metrics.Counter, repository.RepositoryError) {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) StoreGauge(gauge metrics.Gauge) repository.RepositoryError {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) RetrieveGauge(name string) (metrics.Gauge, repository.RepositoryError) {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) ListStoredMetrics() ([]metrics.Gauge, []metrics.Counter, repository.RepositoryError) {
	//TODO implement me
	panic("implement me")
}
