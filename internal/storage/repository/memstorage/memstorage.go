package memstorage

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/KryukovO/metricscollector/internal/metric"
)

type memStorage struct {
	storage []metric.Metrics

	fileStoragePath string
	storeInterval   time.Duration
	tickerDone      chan struct{}
	// TODO: RWMutex для защиты доступа к данным
}

func NewMemStorage(file string, restore bool, storeInterval time.Duration) (*memStorage, error) {
	s := &memStorage{
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
					s.save()
					return
				case <-ticker.C:
					s.save()
				}
			}
		}()
	}

	return s, nil
}

func (s *memStorage) GetAll(ctx context.Context) []metric.Metrics {
	return s.storage
}

func (s *memStorage) GetValue(ctx context.Context, mtype string, mname string) *metric.Metrics {
	for _, mtrc := range s.storage {
		if mtrc.MType == mtype && mtrc.ID == mname {
			return &mtrc
		}
	}
	return nil
}

func (s *memStorage) Update(ctx context.Context, mtrc *metric.Metrics) (err error) {
	defer func() {
		if s.storeInterval == 0 {
			err = s.save()
		}
	}()

	for i := range s.storage {
		if mtrc.MType == s.storage[i].MType && mtrc.ID == s.storage[i].ID {
			if mtrc.Delta != nil {
				*s.storage[i].Delta += *mtrc.Delta
				mtrc.Delta = s.storage[i].Delta
			}
			s.storage[i].Value = mtrc.Value
			return nil
		}
	}

	s.storage = append(s.storage, *mtrc)

	return nil
}

func (s *memStorage) save() error {
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

func (s *memStorage) load() error {
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

func (s *memStorage) Close() error {
	if s.tickerDone != nil {
		s.tickerDone <- struct{}{}
	}
	return nil
}
