package memstorage

import (
	"github.com/KryukovO/metricscollector/internal/metric"
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

func (s *MemStorage) GetAll() map[string]interface{} {
	return s.storage
}

func (s *MemStorage) GetValue(mtype string, mname string) (interface{}, bool) {
	v, ok := s.storage[mname]
	if !ok {
		return nil, ok
	}

	switch mtype {
	case metric.CounterMetric:
		_, ok = v.(int64)
	case metric.GaugeMetric:
		_, ok = v.(float64)
	default:
		ok = false
	}

	if !ok {
		return nil, false
	}

	return v, ok
}

func (s *MemStorage) Update(mtype, mname string, value interface{}) error {
	if mname == "" {
		return storage.ErrWrongMetricName
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
	default:
		return storage.ErrWrongMetricType
	}

	return nil
}
