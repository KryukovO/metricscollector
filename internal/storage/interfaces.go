package storage

import (
	"context"

	"github.com/KryukovO/metricscollector/internal/metric"
)

type Storage interface {
	GetAll(ctx context.Context) ([]metric.Metrics, error)
	GetValue(ctx context.Context, mtype string, mname string) (*metric.Metrics, error)
	Update(ctx context.Context, mtrc *metric.Metrics) error
	UpdateMany(ctx context.Context, mtrc []metric.Metrics) error
	Ping(ctx context.Context) bool
	Close() error
}

type Repo interface {
	GetAll(ctx context.Context) ([]metric.Metrics, error)
	GetValue(ctx context.Context, mtype string, mname string) (*metric.Metrics, error)
	Update(ctx context.Context, mtrc *metric.Metrics) error
	UpdateMany(ctx context.Context, mtrc []metric.Metrics) error
	Ping(ctx context.Context) error
	Close() error
}
