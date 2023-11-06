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
	"github.com/KryukovO/metricscollector/internal/agent/sender"
	"github.com/KryukovO/metricscollector/internal/metric"
	"golang.org/x/sync/errgroup"

	log "github.com/sirupsen/logrus"
)

var ErrServerAddressAnknown = errors.New("unknown server address")

// Agent содержит основные параметры агента.
type Agent struct {
	pollInterval   time.Duration
	reportInterval time.Duration
	sender         sender.Sender
	l              *log.Logger
}

// NewAgent создаёт новый экземпляр структуры агента.
func NewAgent(cfg *config.Config, l *log.Logger) (*Agent, error) {
	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	var (
		snd sender.Sender
		err error
	)

	switch {
	case cfg.GRPCAddress != "":
		snd, err = sender.NewGRPCSender(cfg, lg)
	case cfg.HTTPAddress != "":
		snd, err = sender.NewHTTPSender(cfg, lg)
	default:
		return nil, ErrServerAddressAnknown
	}

	if err != nil {
		return nil, fmt.Errorf("sender initialization error: %w", err)
	}

	return &Agent{
		pollInterval:   cfg.PollInterval.Duration,
		reportInterval: cfg.ReportInterval.Duration,
		sender:         snd,
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
				mtx.Lock()

				sndStorage, err := metricsPreparation(storage, scanCount)
				if err != nil {
					return err
				}

				mtx.Unlock()

				if err = a.sender.Send(ctx, sndStorage); err != nil {
					return err
				}

				a.l.Info("Metrics sender stopped gracefully")

				return nil

			case <-sendTicker.C:
				mtx.Lock()

				sndStorage, err := metricsPreparation(storage, scanCount)
				if err != nil {
					return err
				}

				scanCount = 0

				mtx.Unlock()

				if err = a.sender.Send(ctx, sndStorage); err != nil {
					return err
				}
			}
		}
	})

	return g.Wait()
}

// metricsPreparation выполняет подготовку метрик к отправке на сервер.
func metricsPreparation(storage []metric.Metrics, scanCount int64) ([]metric.Metrics, error) {
	pollCount, err := metric.NewMetrics("PollCount", "", scanCount)
	if err != nil {
		return nil, err
	}

	storage = append(storage, pollCount)
	sndStorage := make([]metric.Metrics, len(storage))

	copy(sndStorage, storage)

	return sndStorage, nil
}
