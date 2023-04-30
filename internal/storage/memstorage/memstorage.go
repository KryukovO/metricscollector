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
		// метрика counter может быть только int64
		val, ok := value.(int64)
		if !ok {
			return storage.ErrWrongMetricValue
		}

		cur, ok := s.storage[mname]
		if ok {
			s.storage[mname] = cur.(int64) + val
		} else {
			s.storage[mname] = value
		}
	case metric.GaugeMetric:
		// метрика gauge может прийти и как float64, и как int64
		val, ok := value.(float64)
		if !ok {
			valInt, ok := value.(int64)
			if !ok {
				return storage.ErrWrongMetricValue
			}
			val = float64(valInt)
		}
		s.storage[mname] = val
	}

	return nil
}
