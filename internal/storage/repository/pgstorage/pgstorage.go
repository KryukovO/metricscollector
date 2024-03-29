// Package pgstorage содержит хранилище метрик, реализованное в репозитории PostgreSQL.
package pgstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/KryukovO/metricscollector/internal/utils"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// PgStorage - хранилище метрик в репозитории PostgreSQL.
type PgStorage struct {
	db      *sql.DB
	retries []int
}

// NewPgStorage - создаёт новое подключение к репозиторию PostgreSQL.
func NewPgStorage(ctx context.Context, dsn, migrations string, retries []int) (*PgStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	s := &PgStorage{
		db:      db,
		retries: retries,
	}

	err = s.Ping(ctx)
	if err != nil {
		return nil, err
	}

	err = s.runMigrations(dsn, migrations)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// runMigrations выполняет миграцию базы данных.
func (s *PgStorage) runMigrations(dsn, migrations string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrations),
		dsn,
	)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

// GetAll возвращает все метрики, находящиеся в репозитории.
func (s *PgStorage) GetAll(ctx context.Context) ([]metric.Metrics, error) {
	slct := func() ([]metric.Metrics, error) {
		query := `
			SELECT 
				mname, mtype, delta, value 
			FROM metrics`

		res := make([]metric.Metrics, 0)

		rows, err := s.db.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}

		defer rows.Close()

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

	var (
		res []metric.Metrics
		err error
	)

	for _, t := range s.retries {
		err = utils.Wait(ctx, time.Duration(t)*time.Second)
		if err != nil {
			return nil, err
		}

		res, err = slct()

		var pgErr *pgconn.PgError
		if err == nil || !errors.As(err, &pgErr) || !pgerrcode.IsConnectionException(pgErr.Code) {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetValue возвращает определенную метрику, соответствующую параметрам mType и mName.
func (s *PgStorage) GetValue(ctx context.Context, mType metric.MetricType, mName string) (*metric.Metrics, error) {
	slct := func() (*metric.Metrics, error) {
		query := `
			SELECT 
				mname, mtype, delta, value 
			FROM metrics
			WHERE mname = $1 AND mtype = $2`

		var (
			delta sql.NullInt64
			value sql.NullFloat64
			mtrc  metric.Metrics
		)

		err := s.db.QueryRowContext(ctx, query, mName, mType).Scan(&mtrc.ID, &mtrc.MType, &delta, &value)
		if err != nil {
			return nil, err
		}

		if delta.Valid {
			mtrc.Delta = &delta.Int64
		} else {
			mtrc.Value = &value.Float64
		}

		return &mtrc, nil
	}

	var (
		mtrc *metric.Metrics
		err  error
	)

	for _, t := range s.retries {
		err = utils.Wait(ctx, time.Duration(t)*time.Second)
		if err != nil {
			return nil, err
		}

		mtrc, err = slct()

		var pgErr *pgconn.PgError
		if err == nil || !errors.As(err, &pgErr) || !pgerrcode.IsConnectionException(pgErr.Code) {
			break
		}
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &metric.Metrics{}, nil
		}

		return nil, err
	}

	return mtrc, nil
}

// Update выполняет обновление единственной метрики.
func (s *PgStorage) Update(ctx context.Context, mtrc *metric.Metrics) error {
	insert := func() (sql.NullInt64, error) {
		query := `
			INSERT INTO metrics(mname, mtype, delta, value) VALUES($1, $2, $3, $4)
			ON CONFLICT (mname, mtype) DO UPDATE SET delta = metrics.delta + $3, value = $4
			RETURNING delta`

		var delta sql.NullInt64

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return sql.NullInt64{}, err
		}

		defer tx.Rollback()

		err = tx.QueryRowContext(ctx, query, mtrc.ID, mtrc.MType, mtrc.Delta, mtrc.Value).Scan(&delta)
		if err != nil {
			return sql.NullInt64{}, err
		}

		return delta, tx.Commit()
	}

	var (
		delta sql.NullInt64
		err   error
	)

	for _, t := range s.retries {
		err = utils.Wait(ctx, time.Duration(t)*time.Second)
		if err != nil {
			return err
		}

		delta, err = insert()

		var pgErr *pgconn.PgError
		if err == nil || !errors.As(err, &pgErr) || !pgerrcode.IsConnectionException(pgErr.Code) {
			break
		}
	}

	if err != nil {
		return err
	}

	if delta.Valid {
		*mtrc.Delta = delta.Int64
	}

	return nil
}

// UpdateMany выполняет обновление метрик из набора.
func (s *PgStorage) UpdateMany(ctx context.Context, mtrcs []metric.Metrics) error {
	insert := func() error {
		query := `
			INSERT INTO metrics(mname, mtype, delta, value) VALUES($1, $2, $3, $4)
			ON CONFLICT (mname, mtype) DO UPDATE SET delta = metrics.delta + $3, value = $4`

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		defer tx.Rollback()

		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return err
		}

		defer stmt.Close()

		for _, mtrc := range mtrcs {
			_, err = stmt.ExecContext(ctx, mtrc.ID, mtrc.MType, mtrc.Delta, mtrc.Value)
			if err != nil {
				return err
			}
		}

		return tx.Commit()
	}

	var err error

	for _, t := range s.retries {
		err = utils.Wait(ctx, time.Duration(t)*time.Second)
		if err != nil {
			return err
		}

		err = insert()

		var pgErr *pgconn.PgError
		if err == nil || !errors.As(err, &pgErr) || !pgerrcode.IsConnectionException(pgErr.Code) {
			break
		}
	}

	return err
}

// Ping выполняет проверку доступности репозитория.
func (s *PgStorage) Ping(ctx context.Context) error {
	var err error
	for _, t := range s.retries {
		err = utils.Wait(ctx, time.Duration(t)*time.Second)
		if err != nil {
			return err
		}

		err = s.db.PingContext(ctx)

		var pgErr *pgconn.PgError
		if err == nil || !errors.As(err, &pgErr) || !pgerrcode.IsConnectionException(pgErr.Code) {
			break
		}
	}

	return err
}

// Close выполняет закрытие репозитория.
func (s *PgStorage) Close() error {
	return s.db.Close()
}
