// Package agent содержит реализацию модуля-агента.
package agent

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/KryukovO/metricscollector/internal/agent/config"
	"github.com/KryukovO/metricscollector/internal/metric"
	"golang.org/x/sync/errgroup"

	log "github.com/sirupsen/logrus"
)

var (
	// ErrStorageIsNil возвращается sender.Send, если был передан неинициализированный слайс метрик.
	ErrStorageIsNil = errors.New("metrics buf is nil")
	// ErrClientIsNil возвращается sender.sendMetrics, если был передан неинициализированный HTTP-клиент.
	ErrClientIsNil = errors.New("HTTP client is nil")
	// ErrUnexpectedStatus возвращается sender.sendMetrics, если сервером был возвращен статус отличный от 200 OK.
	ErrUnexpectedStatus = errors.New("unexpected response status")
)

// Agent содержит основные параметры агента.
type Agent struct {
	pollInterval   time.Duration
	reportInterval time.Duration
	sender         *Sender
	l              *log.Logger
}

// NewAgent создаёт новый экземпляр структуры агента.
func NewAgent(cfg *config.Config, l *log.Logger) (*Agent, error) {
	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	sender, err := NewSender(cfg, lg)
	if err != nil {
		return nil, fmt.Errorf("sender initialization error: %w", err)
	}

	return &Agent{
		pollInterval:   cfg.PollInterval.Duration,
		reportInterval: cfg.ReportInterval.Duration,
		sender:         sender,
		l:              lg,
	}, nil
}

// Run выполняет запуск процессов сканирования и отправки метрик в хранилище.
func (a *Agent) Run(ctx context.Context) error {
	a.l.Info("Agent is running...")

	var (
		scanCount int64
		storage   []metric.Metrics
		mtx       sync.Mutex
		err       error
	)

	scanTicker := time.NewTicker(a.pollInterval)
	sendTicker := time.NewTicker(a.reportInterval)

	sigCtx, sigCancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer sigCancel()

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		for {
			select {
			case <-gCtx.Done():
				return nil

			case <-sigCtx.Done():
				a.l.Info("Metrics scanner stopped gracefully")

				return nil

			case <-scanTicker.C:
				metricCh := ScanMetrics(gCtx)

				mtx.Lock()

				storage = make([]metric.Metrics, 0)

				for mtrc := range metricCh {
					if mtrc.err != nil {
						return err
					}

					storage = append(storage, mtrc.mtrc)
				}

				scanCount++

				mtx.Unlock()
			}
		}
	})

	g.Go(func() error {
		for {
			select {
			case <-gCtx.Done():
				return nil

			case <-sigCtx.Done():
				err = a.send(gCtx, storage, &scanCount, &mtx)
				if err != nil {
					return err
				}

				a.l.Info("Metrics sender stopped gracefully")

				return nil

			case <-sendTicker.C:
				err = a.send(gCtx, storage, &scanCount, &mtx)
				if err != nil {
					return err
				}
			}
		}
	})

	return g.Wait()
}

func (a *Agent) send(ctx context.Context, storage []metric.Metrics, scanCount *int64, mtx *sync.Mutex) error {
	mtx.Lock()

	pollCount, err := metric.NewMetrics("PollCount", "", scanCount)
	if err != nil {
		return err
	}

	storage = append(storage, pollCount)
	*scanCount = 0
	sndStorage := make([]metric.Metrics, len(storage))

	copy(sndStorage, storage)

	mtx.Unlock()

	if err = a.sender.Send(ctx, sndStorage); err != nil {
		return err
	}

	return nil
}
