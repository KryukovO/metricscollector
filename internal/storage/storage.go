package storage

import (
	"errors"

	"github.com/KryukovO/metricscollector/internal/metric"
)

var (
	ErrWrongMetricType  = errors.New("wrong metric type")
	ErrWrongMetricName  = errors.New("wrong metric name")
	ErrWrongMetricValue = errors.New("wrong metric value")
)

type Storage interface {
	GetAll() []metric.Metrics
	GetValue(mtype string, mname string) (*metric.Metrics, bool)
	Update(mtrc *metric.Metrics) error
	Save() error
	Close() error
}
