package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

func NewRepository(dbn string, restore bool) (*Repository, error) {
	db, err := sql.Open("pgx", dbn)
	if err != nil {
		return nil, err
	}
	r := &Repository{
		db: db,
	}

	if !restore {
		r.dropDatabase()
	}

	if err := r.initTable(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Repository) initTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := r.db.ExecContext(ctx, "CREATE TABLE gauges ( name varchar(40) PRIMARY KEY, value double precision NOT NULL)")
	if err != nil {
		logging.Log.S().Errorf("Could not init table: %s", err)
		if !strings.Contains(err.Error(), `relation "gauges" already exists`) {
			return err
		}
	}

	_, err = r.db.ExecContext(ctx, "CREATE TABLE counters ( name varchar(40) PRIMARY KEY, value numeric NOT NULL)")
	if err != nil {
		logging.Log.S().Errorf("Could not init table: %s", err)
		if !strings.Contains(err.Error(), `relation "counters" already exists`) {
			return err
		}
	}

	return nil
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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	c := Counter{
		Name:  counter.Name,
		Value: counter.Value,
	}
	_, err := r.db.ExecContext(ctx, "INSERT INTO counters (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = counters.value + EXCLUDED.value", c.Name, c.Value)
	if err != nil {
		logging.Log.S().Errorf("Could not insert counter: %s", err)
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
		logging.Log.S().Info("Counter does not exits: ", name)
		return metrics.Counter{}, repository.ErrorCounterNotFound

	case err != nil:
		logging.Log.S().Errorf("could not retrieve counter: %s", err)
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
		logging.Log.S().Errorf("Could not insert gauge: %s", err)
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
		logging.Log.S().Info("Counter does not exits: ", name)
		return metrics.Gauge{}, repository.ErrorGaugeNotFound

	case err != nil:
		logging.Log.S().Errorf("could not retrieve counter: %s", err)
		return g.ToMetric(), repository.ErrorGaugeNotFound
	default:
		return g.ToMetric(), nil
	}
}

func (r *Repository) ListStoredMetrics() ([]metrics.Gauge, []metrics.Counter, repository.RepositoryError) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	gauges, err := r.retriveGauges(ctx)
	if err != nil {
		logging.Log.S().Errorf("Error retriving gauges: %s", err)
		return nil, nil, err
	}

	counters, err := r.retriveCounters(ctx)
	if err != nil {
		logging.Log.S().Errorf("Error retriving counters: %s", err)
		return nil, nil, err
	}

	return gauges, counters, nil
}

func (r Repository) retriveCounters(ctx context.Context) ([]metrics.Counter, error) {
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
func (r Repository) retriveGauges(ctx context.Context) ([]metrics.Gauge, error) {
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
