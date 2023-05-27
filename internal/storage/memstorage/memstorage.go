package memstorage

import (
	"encoding/json"
	"os"
	"time"

	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/KryukovO/metricscollector/internal/storage"
)

type MemStorage struct {
	storage []metric.Metrics

	fileStoragePath string
	storeInterval   time.Duration
	tickerDone      chan struct{}
}

func NewMemStorage(file string, restore bool, storeInterval time.Duration) (*MemStorage, error) {
	s := &MemStorage{
		storage:         make([]metric.Metrics, 0),
		fileStoragePath: file,
		storeInterval:   storeInterval,
	}
	if restore {
		err := s.load()
		if err != nil {
			return nil, err
		}
	}
	if storeInterval > 0 {
		s.tickerDone = make(chan struct{})
		ticker := time.NewTicker(storeInterval)
		go func() {
			for {
				select {
				case <-s.tickerDone:
					s.Save()
					return
				case <-ticker.C:
					s.Save()
				}
			}
		}()
	}

	return s, nil
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

func (s *MemStorage) Update(mtrc *metric.Metrics) (err error) {
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

	defer func() {
		if s.storeInterval == 0 {
			err = s.Save()
		}
	}()

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

func (s *MemStorage) Save() error {
	if s.fileStoragePath != "" {
		file, err := os.OpenFile(s.fileStoragePath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer file.Close()
		encoder := json.NewEncoder(file)
		return encoder.Encode(&s.storage)
	}
	return nil
}

func (s *MemStorage) load() error {
	if s.fileStoragePath != "" {
		file, err := os.OpenFile(s.fileStoragePath, os.O_RDONLY, 0666)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		defer file.Close()
		decoder := json.NewDecoder(file)
		return decoder.Decode(&s.storage)
	}
	return nil
}

func (s *MemStorage) Close() error {
	if s.tickerDone != nil {
		s.tickerDone <- struct{}{}
	}
	return nil
}
