// Package memstorage содержит in-memory хранилище метрик.
package memstorage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/KryukovO/metricscollector/internal/utils"
	log "github.com/sirupsen/logrus"
)

// MemStorage - хранилище метрик с репозиторием в памяти сервера.
// Репозиторий поддерживает функциональность сброса содержимого в файл на сервере.
type MemStorage struct {
	storage []metric.Metrics // in-memory хранилище метрик

	fileStoragePath string // путь до файла, в который сохраняются метрики
	syncSave        bool   // признак синхронной записи в файл
	closeSave       func() // функция, закрывающая горутину, которая пишет в файл
	retries         []int
	mtx             sync.RWMutex
	l               *log.Logger
}

// NewMemStorage создаёт новое in-memory хранилище.
func NewMemStorage(
	ctx context.Context, file string, restore bool,
	storeInterval time.Duration, retries []int, l *log.Logger,
) (*MemStorage, error) {
	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	s := &MemStorage{
		storage:         make([]metric.Metrics, 0),
		fileStoragePath: file,
		retries:         retries,
		syncSave:        storeInterval == 0,
		closeSave:       func() {},
		l:               lg,
	}

	if restore {
		err := s.load(ctx)
		if err != nil {
			return nil, err
		}
	}

	if file != "" && storeInterval > 0 {
		saveCtx, cancel := context.WithCancel(context.Background())
		s.closeSave = cancel
		ticker := time.NewTicker(storeInterval)

		go func() {
			for {
				select {
				case <-saveCtx.Done():
					ticker.Stop()

					return

				case <-ticker.C:
					if err := s.save(saveCtx); err != nil {
						s.l.Errorf("error when saving metrics to the file: %s", err)
					}
				}
			}
		}()
	}

	return s, nil
}

// update выполняет обновление метрики в репозитории.
func (s *MemStorage) update(mtrc *metric.Metrics) {
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

// save выполняет сохранение метрик из памяти сервера в файл.
func (s *MemStorage) save(ctx context.Context) error {
	const filePerm fs.FileMode = 0o666

	var (
		file *os.File
		err  error
	)

	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.fileStoragePath != "" {
		for _, t := range s.retries {
			err = utils.Wait(ctx, time.Duration(t)*time.Second)
			if err != nil {
				return err
			}

			file, err = os.OpenFile(s.fileStoragePath, os.O_WRONLY|os.O_CREATE, filePerm)
			if err == nil || !errors.Is(err, syscall.EBUSY) {
				break
			}
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

// load выполняет загрузку метрик из файла в память сервера.
func (s *MemStorage) load(ctx context.Context) error {
	var (
		data []byte
		err  error
	)

	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.fileStoragePath == "" {
		return nil
	}

	for _, t := range s.retries {
		err = utils.Wait(ctx, time.Duration(t)*time.Second)
		if err != nil {
			return err
		}

		data, err = os.ReadFile(s.fileStoragePath)
		if err == nil || !errors.Is(err, syscall.EBUSY) {
			break
		}
	}

	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	decoder := json.NewDecoder(bytes.NewReader(data))

	return decoder.Decode(&s.storage)
}

// GetAll возвращает все метрики, находящиеся в репозитории.
func (s *MemStorage) GetAll(_ context.Context) ([]metric.Metrics, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	data := make([]metric.Metrics, len(s.storage))
	copy(data, s.storage)

	return data, nil
}

// GetValue возвращает определенную метрику, соответствующую параметрам mType и mName.
func (s *MemStorage) GetValue(_ context.Context, mType string, mName string) (*metric.Metrics, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	for _, mtrc := range s.storage {
		if mtrc.MType == mType && mtrc.ID == mName {
			return &mtrc, nil
		}
	}

	return &metric.Metrics{}, nil
}

// Update выполняет обновление единственной метрики.
func (s *MemStorage) Update(ctx context.Context, mtrc *metric.Metrics) error {
	defer func() {
		if s.syncSave {
			err := s.save(ctx)
			if err != nil {
				s.l.Errorf("error when saving metrics to the file: %s", err)
			}
		}
	}()

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.update(mtrc)

	return nil
}

// UpdateMany выполняет обновление метрик из набора.
func (s *MemStorage) UpdateMany(_ context.Context, mtrcs []metric.Metrics) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for i := 0; i < len(mtrcs); i++ {
		s.update(&mtrcs[i])
	}

	return nil
}

// Ping выполняет проверку доступности репозитория.
func (s *MemStorage) Ping(_ context.Context) error {
	return nil
}

// Close выполняет закрытие репозитория.
func (s *MemStorage) Close() error {
	s.closeSave()

	// Сохранение метрик при закрытии хранилища
	if err := s.save(context.Background()); err != nil {
		s.l.Errorf("error when saving metrics to the file: %s", err)
	}

	return nil
}
