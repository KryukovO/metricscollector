package storage

import (
	"context"
	"errors"

	"github.com/KryukovO/metricscollector/internal/metric"
)

var (
	ErrWrongMetricType  = errors.New("wrong metric type")
	ErrWrongMetricName  = errors.New("wrong metric name")
	ErrWrongMetricValue = errors.New("wrong metric value")
)

type Storage interface {
	GetAll(ctx context.Context) []metric.Metrics
	GetValue(ctx context.Context, mtype string, mname string) (*metric.Metrics, bool)
	Update(ctx context.Context, mtrc *metric.Metrics) error
	Close() error
}

type StorageRepo interface {
	GetAll(ctx context.Context) []metric.Metrics
	GetValue(ctx context.Context, mtype string, mname string) *metric.Metrics
	Update(ctx context.Context, mtrc *metric.Metrics) error
	Close() error
}
