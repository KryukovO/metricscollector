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
	return s.repo.GetValue(ctx, mtype, mname)
}

func (s *storage) Update(ctx context.Context, mtrc *metric.Metrics) error {
	if mtrc.ID == "" {
		return ErrWrongMetricName
	}

	switch mtrc.MType {
	case metric.CounterMetric:
		if mtrc.Delta == nil {
			return ErrWrongMetricValue
		}
		mtrc.Value = nil
	case metric.GaugeMetric:
		if mtrc.Value == nil {
			return ErrWrongMetricValue
		}
		mtrc.Delta = nil
	default:
		return ErrWrongMetricType
	}

	return s.repo.Update(ctx, mtrc)
}

func (s *storage) Ping() bool {
	return s.repo.Ping() == nil
}

func (s *storage) Close() error {
	return s.repo.Close()
}
