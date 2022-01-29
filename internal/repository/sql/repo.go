package sql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/OmAsana/yapraktikum/internal/logging"
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
	r := &Repository{
		db: db,
	}
	if err := r.initTable(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Repository) initTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	result, err := r.db.ExecContext(ctx, "CREATE TABLE gauges { name varchar(40), value integer NOT NULL, CONSTRAINT prod UNIQUE(name)}")
	if err != nil {
		logging.Log.S().Errorf("Could not init table: %s", err)
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		logging.Log.S().Errorf("Got err creating table: %s", err)
		return err
	}

	logging.Log.S().Info("Rows affecred: ", n)

}

func (r *Repository) Ping() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := r.db.PingContext(ctx)
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
