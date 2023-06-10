package memstorage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/KryukovO/metricscollector/internal/metric"
	log "github.com/sirupsen/logrus"
)

type memStorage struct {
	storage []metric.Metrics // in-memory хранилище метрик

	fileStoragePath string        // путь до файла, в который сохраняются метрики
	syncSave        bool          // признак синхронной записи в файл
	closeSaving     chan struct{} // канал, принимающий сообщение о необходимости прекратить сохранения в файл
	mtx             sync.RWMutex
	l               *log.Logger
}

func NewMemStorage(file string, restore bool, storeInterval time.Duration, l *log.Logger) (*memStorage, error) {
	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	s := &memStorage{
		storage:         make([]metric.Metrics, 0),
		fileStoragePath: file,
		syncSave:        storeInterval == 0,
		l:               lg,
	}
	if restore {
		err := s.load()
		if err != nil {
			return nil, err
		}
	}
	if file != "" && storeInterval > 0 {
		s.closeSaving = make(chan struct{})
		ticker := time.NewTicker(storeInterval)
		go func() {
			for {
				select {
				case <-s.closeSaving:
					err := s.save()
					if err != nil {
						s.l.Infof("error when saving metrics to the file: %s", err)
					}
					return
				case <-ticker.C:
					err := s.save()
					if err != nil {
						s.l.Infof("error when saving metrics to the file: %s", err)
					}
				}
			}
		}()
	}

	return s, nil
}

func (s *memStorage) update(mtrc *metric.Metrics) {
	for i := range s.storage {
		if mtrc.MType == s.storage[i].MType && mtrc.ID == s.storage[i].ID {
			if mtrc.Delta != nil {
				*s.storage[i].Delta += *mtrc.Delta
				mtrc.Delta = s.storage[i].Delta
			}
			s.storage[i].Value = mtrc.Value
			return
		}
	}

	s.storage = append(s.storage, *mtrc)
}

func (s *memStorage) save() error {
	var (
		file *os.File
		err  error
	)

	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.fileStoragePath != "" {
		for t := 1; t <= 5; t += 2 {
			file, err = os.OpenFile(s.fileStoragePath, os.O_WRONLY|os.O_CREATE, 0666)
			if err == nil || !errors.Is(err, syscall.EBUSY) {
				break
			}
			time.Sleep(time.Duration(t) * time.Second)
		}
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
	var (
		file *os.File
		err  error
	)

	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.fileStoragePath != "" {
		for t := 1; t <= 5; t += 2 {
			file, err = os.OpenFile(s.fileStoragePath, os.O_RDONLY, 0666)
			if err == nil || !errors.Is(err, syscall.EBUSY) {
				break
			}
			time.Sleep(time.Duration(t) * time.Second)
		}
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

func (s *memStorage) GetAll(ctx context.Context) ([]metric.Metrics, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	data := make([]metric.Metrics, len(s.storage))
	copy(data, s.storage)
	return data, nil
}

func (s *memStorage) GetValue(ctx context.Context, mtype string, mname string) (*metric.Metrics, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	for _, mtrc := range s.storage {
		if mtrc.MType == mtype && mtrc.ID == mname {
			return &mtrc, nil
		}
	}
	return nil, nil
}

func (s *memStorage) Update(ctx context.Context, mtrc *metric.Metrics) error {
	defer func() {
		if s.syncSave {
			err := s.save()
			if err != nil {
				s.l.Infof("error when saving metrics to the file: %s", err)
			}
		}
	}()

	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.update(mtrc)

	return nil
}

func (s *memStorage) UpdateMany(ctx context.Context, mtrcs []metric.Metrics) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	for _, mtrc := range mtrcs {
		s.update(&mtrc)
	}

	return nil
}

func (s *memStorage) Ping() error {
	return nil
}

func (s *memStorage) Close() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if s.closeSaving != nil {
		s.closeSaving <- struct{}{}
		s.closeSaving = nil
	}
	return nil
}
