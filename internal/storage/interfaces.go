package storage

import (
	"context"

	"github.com/KryukovO/metricscollector/internal/metric"
)

// Storage - интерфейс логики взаимодействия с хранилищем.
type Storage interface {
	// GetAll возвращает все метрики, находящиеся в хранилище.
	GetAll(ctx context.Context) ([]metric.Metrics, error)
	// GetValue возвращает определенную метрику, соответствующую параметрам mType и mName.
	GetValue(ctx context.Context, mType metric.MetricType, mName string) (*metric.Metrics, error)
	// Update выполняет обновление единственной метрики.
	Update(ctx context.Context, mtrc *metric.Metrics) error
	// UpdateMany выполняет обновление метрик из набора.
	UpdateMany(ctx context.Context, mtrc []metric.Metrics) error
	// Ping выполняет проверку доступности хранилища.
	Ping(ctx context.Context) bool
	// Close выполняет закрытие хранилища.
	Close() error
}

// Repo - интерфейс взаимодейтсвия с репозиторием.
type Repo interface {
	// GetAll возвращает все метрики, находящиеся в репозитории.
	GetAll(ctx context.Context) ([]metric.Metrics, error)
	// GetValue возвращает определенную метрику, соответствующую параметрам mType и mName.
	GetValue(ctx context.Context, mType metric.MetricType, mName string) (*metric.Metrics, error)
	// Update выполняет обновление единственной метрики.
	Update(ctx context.Context, mtrc *metric.Metrics) error
	// UpdateMany выполняет обновление метрик из набора.
	UpdateMany(ctx context.Context, mtrc []metric.Metrics) error
	// Ping выполняет проверку доступности репозитория.
	Ping(ctx context.Context) error
	// Close выполняет закрытие репозитория.
	Close() error
}
