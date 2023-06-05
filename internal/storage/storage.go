package storage

import (
	"context"

	"github.com/KryukovO/metricscollector/internal/metric"
)

type storage struct {
	repo StorageRepo
}

func NewStorage(repo StorageRepo) *storage {
	return &storage{
		repo: repo,
	}
}

func (s *storage) GetAll(ctx context.Context) ([]metric.Metrics, error) {
	return s.repo.GetAll(ctx)
}

func (s *storage) GetValue(ctx context.Context, mtype string, mname string) (*metric.Metrics, error) {
	if mtype != metric.CounterMetric && mtype != metric.GaugeMetric {
		return nil, metric.ErrWrongMetricType
	}

	return s.repo.GetValue(ctx, mtype, mname)
}

func (s *storage) Update(ctx context.Context, mtrc *metric.Metrics) error {
	if err := mtrc.Validate(); err != nil {
		return err
	}
	return s.repo.Update(ctx, mtrc)
}

func (s *storage) UpdateMany(ctx context.Context, mtrcs []metric.Metrics) error {
	for _, mtrc := range mtrcs {
		if err := mtrc.Validate(); err != nil {
			return err
		}
	}
	return s.repo.UpdateMany(ctx, mtrcs)
}

func (s *storage) Ping() bool {
	return s.repo.Ping() == nil
}

func (s *storage) Close() error {
	return s.repo.Close()
}
