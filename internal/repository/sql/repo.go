package sql

import (
	"context"
	"database/sql"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/OmAsana/yapraktikum/internal/logging"
	"github.com/OmAsana/yapraktikum/internal/metrics"
	"github.com/OmAsana/yapraktikum/internal/repository"
)

var _ repository.MetricsRepository = (*Repository)(nil)

type Repository struct {
	db  *sql.DB
	log *logging.Logger
}

func (r *Repository) WriteBulkCounters(counters []metrics.Counter) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO counters (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = counters.value + EXCLUDED.value")

	if err != nil {
		return err
	}

	for _, v := range counters {
		if _, err := stmt.Exec(v.Name, v.Value); err != nil {
			if err = tx.Rollback(); err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *Repository) WriteBulkGauges(gauges []metrics.Gauge) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO gauges (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE set VALUE = EXCLUDED.value")

	if err != nil {
		return err
	}

	for _, v := range gauges {
		if _, err := stmt.Exec(v.Name, v.Value); err != nil {
			if err = tx.Rollback(); err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func NewRepository(dbn string, restore bool, opts ...Option) (*Repository, error) {
	db, err := sql.Open("pgx", dbn)
	if err != nil {
		return nil, err
	}
	r := &Repository{
		db:  db,
		log: logging.NewNoop(),
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	if !restore {
		if err := r.dropDatabase(); err != nil {
			return nil, err
		}
	}

	if err := r.initTable(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Repository) initTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := r.db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS gauges ( name varchar(40) PRIMARY KEY, value double precision NOT NULL)")
	if err != nil {
		r.log.S().Errorf("Could not init table: %s", err)
		if !strings.Contains(err.Error(), `relation "gauges" already exists`) {
			return err
		}
	}

	_, err = r.db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS counters ( name varchar(40) PRIMARY KEY, value numeric NOT NULL)")
	if err != nil {
		r.log.S().Errorf("Could not init table: %s", err)
		if !strings.Contains(err.Error(), `relation "counters" already exists`) {
			return err
		}
	}

	return nil
}

func (r *Repository) Ping() bool {
	err := r.db.Ping()
	return err == nil
}

func (r *Repository) StoreCounter(counter metrics.Counter) repository.RepositoryError {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	c := Counter{
		Name:  counter.Name,
		Value: counter.Value,
	}
	_, err := r.db.ExecContext(ctx, "INSERT INTO counters (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = counters.value + EXCLUDED.value", c.Name, c.Value)
	if err != nil {
		r.log.S().Errorf("Could not insert counter: %s", err)
		return repository.ErrorInternalError
	}
	return nil
}

func (r *Repository) RetrieveCounter(name string) (metrics.Counter, repository.RepositoryError) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	sqlStatement := `SELECT name, value from counters where name=$1`
	c := Counter{}
	err := r.db.QueryRowContext(ctx, sqlStatement, name).Scan(&c.Name, &c.Value)
	switch {
	case err == sql.ErrNoRows:
		r.log.S().Info("Counter does not exits: ", name)
		return metrics.Counter{}, repository.ErrorCounterNotFound

	case err != nil:
		r.log.S().Errorf("could not retrieve counter: %s", err)
		return c.ToMetric(), repository.ErrorCounterNotFound
	default:
		return c.ToMetric(), nil
	}

}

func (r *Repository) StoreGauge(gauge metrics.Gauge) repository.RepositoryError {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	g := Gauge{
		Name:  gauge.Name,
		Delta: gauge.Value,
	}
	_, err := r.db.ExecContext(ctx, "INSERT INTO gauges (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE set VALUE = EXCLUDED.value", g.Name, g.Delta)
	if err != nil {
		r.log.S().Errorf("Could not insert gauge: %s", err)
		return repository.ErrorInternalError
	}
	return nil
}

func (r *Repository) RetrieveGauge(name string) (metrics.Gauge, repository.RepositoryError) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	sqlStatement := `SELECT name, value from gauges where name=$1`
	g := Gauge{}
	err := r.db.QueryRowContext(ctx, sqlStatement, name).Scan(&g.Name, &g.Delta)
	switch {
	case err == sql.ErrNoRows:
		r.log.S().Info("Counter does not exits: ", name)
		return metrics.Gauge{}, repository.ErrorGaugeNotFound

	case err != nil:
		r.log.S().Errorf("could not retrieve counter: %s", err)
		return g.ToMetric(), repository.ErrorGaugeNotFound
	default:
		return g.ToMetric(), nil
	}
}

func (r *Repository) ListStoredMetrics() ([]metrics.Gauge, []metrics.Counter, repository.RepositoryError) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	gauges, err := r.retrieveGauges(ctx)
	if err != nil {
		r.log.S().Errorf("Error retriving gauges: %s", err)
		return nil, nil, err
	}

	counters, err := r.retrieveCounters(ctx)
	if err != nil {
		r.log.S().Errorf("Error retriving counters: %s", err)
		return nil, nil, err
	}

	return gauges, counters, nil
}

func (r Repository) retrieveCounters(ctx context.Context) ([]metrics.Counter, error) {
	var counters []metrics.Counter
	sqlStatement := `SELECT * FROM counters`

	rows, err := r.db.QueryContext(ctx, sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var counter Counter
		if err := rows.Scan(&counter.Name, &counter.Value); err != nil {
			return nil, err
		}
		counters = append(counters, counter.ToMetric())
	}

	rerr := rows.Close()
	if rerr != nil {
		return nil, err
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return counters, nil
}
func (r Repository) retrieveGauges(ctx context.Context) ([]metrics.Gauge, error) {
	var gauges []metrics.Gauge
	sqlStatement := `SELECT * FROM gauges`

	rows, err := r.db.QueryContext(ctx, sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var gauge Gauge
		if err := rows.Scan(&gauge.Name, &gauge.Delta); err != nil {
			return nil, err
		}
		gauges = append(gauges, gauge.ToMetric())

	}

	rerr := rows.Close()
	if rerr != nil {
		return nil, err
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return gauges, nil
}

func (r *Repository) dropDatabase() error {
	sqlStatement := `DROP TABLE IF EXISTS counters, gauges CASCADE`
	_, err := r.db.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}
