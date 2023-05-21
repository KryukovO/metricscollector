package memstorage

import (
	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/KryukovO/metricscollector/internal/storage"
)

type MemStorage struct {
	storage []metric.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		storage: make([]metric.Metrics, 0),
	}
}

func (s *MemStorage) GetAll() []metric.Metrics {
	return s.storage
}

func (s *MemStorage) GetValue(mtype string, mname string) (*metric.Metrics, bool) {
	for _, mtrc := range s.storage {
		if mtrc.MType == mtype && mtrc.ID == mname {
			return &mtrc, true
		}
	}
	return nil, false
}

func (s *MemStorage) Update(mtrc *metric.Metrics) error {
	if mtrc.ID == "" {
		return storage.ErrWrongMetricName
	}

	var (
		counterVal *int64
		gaugeVal   *float64
	)

	switch mtrc.MType {
	case metric.CounterMetric:
		if mtrc.Delta == nil {
			return storage.ErrWrongMetricValue
		}
		counterVal = mtrc.Delta
	case metric.GaugeMetric:
		if mtrc.Value == nil {
			return storage.ErrWrongMetricValue
		}
		gaugeVal = mtrc.Value
	default:
		return storage.ErrWrongMetricType
	}

	for i := range s.storage {
		if mtrc.MType == s.storage[i].MType && mtrc.ID == s.storage[i].ID {
			if counterVal != nil {
				*s.storage[i].Delta += *counterVal
				mtrc.Delta = s.storage[i].Delta
			}
			s.storage[i].Value = gaugeVal
			return nil
		}
	}

	s.storage = append(s.storage, metric.Metrics{
		ID:    mtrc.ID,
		MType: mtrc.MType,
		Delta: counterVal,
		Value: gaugeVal,
	})

	return nil
}
