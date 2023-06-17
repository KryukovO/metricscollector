package storage

import (
	"context"
	"time"

	"github.com/KryukovO/metricscollector/internal/metric"
)

type MetricsStorage struct {
	repo    Repo
	timeout time.Duration
}

func NewMetricsStorage(repo Repo, timeout uint) *MetricsStorage {
	return &MetricsStorage{
		repo:    repo,
		timeout: time.Duration(timeout) * time.Second,
	}
}

func (s *MetricsStorage) GetAll(ctx context.Context) ([]metric.Metrics, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.repo.GetAll(ctx)
}

func (s *MetricsStorage) GetValue(ctx context.Context, mType string, mName string) (*metric.Metrics, error) {
	if mType != metric.CounterMetric && mType != metric.GaugeMetric {
		return nil, metric.ErrWrongMetricType
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.repo.GetValue(ctx, mType, mName)
}

func (s *MetricsStorage) Update(ctx context.Context, mtrc *metric.Metrics) error {
	if err := mtrc.Validate(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.repo.Update(ctx, mtrc)
}

func (s *MetricsStorage) UpdateMany(ctx context.Context, mtrcs []metric.Metrics) error {
	for _, mtrc := range mtrcs {
		if err := mtrc.Validate(); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.repo.UpdateMany(ctx, mtrcs)
}

func (s *MetricsStorage) Ping(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.repo.Ping(ctx) == nil
}

func (s *MetricsStorage) Close() error {
	return s.repo.Close()
}
