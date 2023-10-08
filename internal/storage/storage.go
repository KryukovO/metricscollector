// Package storage содержит инструментарий для взаимодействия с хранилищем метрик.
package storage

import (
	"context"
	"time"

	"github.com/KryukovO/metricscollector/internal/metric"
)

// MetricsStorage структура, обеспечивающая взаимодействие с хранилищем.
type MetricsStorage struct {
	repo    Repo
	timeout time.Duration
}

// NewMetricsStorage создаёт новую структуру для взаимодействия с хранилищем.
func NewMetricsStorage(repo Repo, timeout uint) *MetricsStorage {
	return &MetricsStorage{
		repo:    repo,
		timeout: time.Duration(timeout) * time.Second,
	}
}

// GetAll возвращает все метрики, находящиеся в хранилище.
func (s *MetricsStorage) GetAll(ctx context.Context) ([]metric.Metrics, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.repo.GetAll(ctx)
}

// GetValue возвращает определенную метрику, соответствующую параметрам mType и mName.
func (s *MetricsStorage) GetValue(ctx context.Context, mType string, mName string) (*metric.Metrics, error) {
	if mType != metric.CounterMetric && mType != metric.GaugeMetric {
		return nil, metric.ErrWrongMetricType
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.repo.GetValue(ctx, mType, mName)
}

// Update выполняет обновление единственной метрики.
func (s *MetricsStorage) Update(ctx context.Context, mtrc *metric.Metrics) error {
	if err := mtrc.Validate(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.repo.Update(ctx, mtrc)
}

// UpdateMany выполняет обновление метрик из набора.
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

// Ping выполняет проверку доступности хранилища.
func (s *MetricsStorage) Ping(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.repo.Ping(ctx) == nil
}

// Close выполняет закрытие хранилища.
func (s *MetricsStorage) Close() error {
	return s.repo.Close()
}
