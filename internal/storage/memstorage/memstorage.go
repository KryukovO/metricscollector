package memstorage

import (
	"github.com/KryukovO/metricscollector/internal/models/metric"
	"github.com/KryukovO/metricscollector/internal/storage"
)

type MemStorage struct {
	storage map[string]interface{}
}

func New() *MemStorage {
	return &MemStorage{
		storage: make(map[string]interface{}),
	}
}

func (s *MemStorage) Update(mtype, mname string, value interface{}) error {
	if mtype != metric.CounterMetric && mtype != metric.GaugeMetric {
		return storage.ErrWrongMetricType
	}

	switch mtype {
	case metric.CounterMetric:
		cur, ok := s.storage[mname]
		if ok {
			s.storage[mname] = cur.(int64) + value.(int64)
		} else {
			s.storage[mname] = value
		}
	case metric.GaugeMetric:
		s.storage[mname] = value
	}

	return nil
}
