package storage

import (
	"context"

	"github.com/KryukovO/metricscollector/internal/metric"
)

type MetricsStorage struct {
	repo Repo
}

func NewMetricsStorage(repo Repo) *MetricsStorage {
	return &MetricsStorage{
		repo: repo,
	}
}

func (s *MetricsStorage) GetAll(ctx context.Context) ([]metric.Metrics, error) {
	return s.repo.GetAll(ctx)
}

func (s *MetricsStorage) GetValue(ctx context.Context, mtype string, mname string) (*metric.Metrics, error) {
	if mtype != metric.CounterMetric && mtype != metric.GaugeMetric {
		return nil, metric.ErrWrongMetricType
	}

	return s.repo.GetValue(ctx, mtype, mname)
}

func (s *MetricsStorage) Update(ctx context.Context, mtrc *metric.Metrics) error {
	if err := mtrc.Validate(); err != nil {
		return err
	}

	return s.repo.Update(ctx, mtrc)
}

func (s *MetricsStorage) UpdateMany(ctx context.Context, mtrcs []metric.Metrics) error {
	for _, mtrc := range mtrcs {
		if err := mtrc.Validate(); err != nil {
			return err
		}
	}

	return s.repo.UpdateMany(ctx, mtrcs)
}

func (s *MetricsStorage) Ping() bool {
	return s.repo.Ping() == nil
}

func (s *MetricsStorage) Close() error {
	return s.repo.Close()
}
