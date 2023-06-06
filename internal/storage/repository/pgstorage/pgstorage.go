package pgstorage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type pgStorage struct {
	db      *sql.DB
	timeout int
}

func NewPgStorage(dsn string) (*pgStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	s := &pgStorage{db: db, timeout: 2}
	err = s.upSchema()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *pgStorage) upSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.timeout)*time.Second)
	defer cancel()

	query := `
		CREATE TABLE IF NOT EXISTS metrics(
			id TEXT NOT NULL,
			mtype TEXT NOT NULL,
			delta BIGINT,
			value DOUBLE PRECISION,
			PRIMARY KEY(id, mtype)
		);`

	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (s *pgStorage) GetAll(ctx context.Context) ([]metric.Metrics, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.timeout)*time.Second)
	defer cancel()

	query := `
		SELECT 
			id, mtype, delta, value 
		FROM metrics`

	var (
		rows *sql.Rows
		err  error
	)

	for t := 1; t <= 5; t += 2 {
		rows, err = s.db.QueryContext(ctx, query)
		var pgErr *pgconn.PgError
		if err == nil || !errors.As(err, &pgErr) || !pgerrcode.IsConnectionException(pgErr.Code) {
			break
		}
		time.Sleep(time.Duration(t) * time.Second)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]metric.Metrics, 0)
	for rows.Next() {
		var (
			delta sql.NullInt64
			value sql.NullFloat64
			mtrc  metric.Metrics
		)

		err = rows.Scan(&mtrc.ID, &mtrc.MType, &delta, &value)
		if err != nil {
			return nil, err
		}

		if delta.Valid {
			mtrc.Delta = &delta.Int64
		} else {
			mtrc.Value = &value.Float64
		}

		res = append(res, mtrc)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *pgStorage) GetValue(ctx context.Context, mtype string, mname string) (*metric.Metrics, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.timeout)*time.Second)
	defer cancel()

	query := `
		SELECT 
			id, mtype, delta, value 
		FROM metrics
		WHERE id = $1 AND mtype = $2`

	var (
		delta sql.NullInt64
		value sql.NullFloat64
		mtrc  metric.Metrics
		err   error
	)

	for t := 1; t <= 5; t += 2 {
		err = s.db.QueryRowContext(ctx, query, mname, mtype).Scan(&mtrc.ID, &mtrc.MType, &delta, &value)
		var pgErr *pgconn.PgError
		if err == nil || !errors.As(err, &pgErr) || !pgerrcode.IsConnectionException(pgErr.Code) {
			break
		}
		time.Sleep(time.Duration(t) * time.Second)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if delta.Valid {
		mtrc.Delta = &delta.Int64
	} else {
		mtrc.Value = &value.Float64
	}

	return &mtrc, nil
}

func (s *pgStorage) Update(ctx context.Context, mtrc *metric.Metrics) (err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.timeout)*time.Second)
	defer cancel()

	query := `
		INSERT INTO metrics(id, mtype, delta, value) VALUES($1, $2, $3, $4)
		ON CONFLICT (id, mtype) DO UPDATE SET delta = metrics.delta + $3, value = $4`

	for t := 1; t <= 5; t += 2 {
		_, err = s.db.ExecContext(ctx, query, mtrc.ID, mtrc.MType, mtrc.Delta, mtrc.Value)
		var pgErr *pgconn.PgError
		if err == nil || !errors.As(err, &pgErr) || !pgerrcode.IsConnectionException(pgErr.Code) {
			break
		}
		time.Sleep(time.Duration(t) * time.Second)
	}

	return
}

func (s *pgStorage) UpdateMany(ctx context.Context, mtrcs []metric.Metrics) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.timeout)*time.Second)
	defer cancel()

	query := `
		INSERT INTO metrics(id, mtype, delta, value) VALUES($1, $2, $3, $4)
		ON CONFLICT (id, mtype) DO UPDATE SET delta = metrics.delta + $3, value = $4`

	var (
		tx   *sql.Tx
		stmt *sql.Stmt
		err  error
	)

	for t := 1; t <= 5; t += 2 {
		tx, err = s.db.BeginTx(ctx, nil)
		var pgErr *pgconn.PgError
		if err == nil || !errors.As(err, &pgErr) || !pgerrcode.IsConnectionException(pgErr.Code) {
			break
		}
		time.Sleep(time.Duration(t) * time.Second)
	}
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for t := 1; t <= 5; t += 2 {
		stmt, err = tx.PrepareContext(ctx, query)
		var pgErr *pgconn.PgError
		if err == nil || !errors.As(err, &pgErr) || !pgerrcode.IsConnectionException(pgErr.Code) {
			break
		}
		time.Sleep(time.Duration(t) * time.Second)
	}
	if err != nil {
		return err
	}

	for _, mtrc := range mtrcs {
		for t := 1; t <= 5; t += 2 {
			_, err := stmt.ExecContext(ctx, mtrc.ID, mtrc.MType, mtrc.Delta, mtrc.Value)
			var pgErr *pgconn.PgError
			if err == nil || !errors.As(err, &pgErr) || !pgerrcode.IsConnectionException(pgErr.Code) {
				break
			}
			time.Sleep(time.Duration(t) * time.Second)
		}
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *pgStorage) Ping() (err error) {
	for t := 1; t <= 5; t += 2 {
		err = s.db.Ping()
		var pgErr *pgconn.PgError
		if err == nil || !errors.As(err, &pgErr) || !pgerrcode.IsConnectionException(pgErr.Code) {
			break
		}
		time.Sleep(time.Duration(t) * time.Second)
	}
	return
}

func (s *pgStorage) Close() error {
	return s.db.Close()
}
