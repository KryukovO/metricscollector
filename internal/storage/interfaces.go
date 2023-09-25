package storage

import (
	"context"

	"github.com/KryukovO/metricscollector/internal/metric"
)

// Интерфейс логики взаимодействия с хранилищем.
type Storage interface {
	// Возвращает все метрики, находящиеся в хранилище.
	GetAll(ctx context.Context) ([]metric.Metrics, error)
	// Возвращает определенную метрику, соответствующую параметрам mType и mName.
	GetValue(ctx context.Context, mType string, mName string) (*metric.Metrics, error)
	// Выполняет обновление единственной метрики.
	Update(ctx context.Context, mtrc *metric.Metrics) error
	// Выполняет обновление метрик из набора.
	UpdateMany(ctx context.Context, mtrc []metric.Metrics) error
	// Выполняет проверку доступности хранилища.
	Ping(ctx context.Context) bool
	// Выполняет закрытие хранилища.
	Close() error
}

// Интерфейс взаимодейтсвия с репозиторием.
type Repo interface {
	// Возвращает все метрики, находящиеся в репозитории.
	GetAll(ctx context.Context) ([]metric.Metrics, error)
	// Возвращает определенную метрику, соответствующую параметрам mType и mName.
	GetValue(ctx context.Context, mType string, mName string) (*metric.Metrics, error)
	// Выполняет обновление единственной метрики.
	Update(ctx context.Context, mtrc *metric.Metrics) error
	// Выполняет обновление метрик из набора.
	UpdateMany(ctx context.Context, mtrc []metric.Metrics) error
	// Выполняет проверку доступности репозитория.
	Ping(ctx context.Context) error
	// Выполняет закрытие репозитория.
	Close() error
}
