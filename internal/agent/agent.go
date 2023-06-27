package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/KryukovO/metricscollector/internal/agent/config"
	"github.com/KryukovO/metricscollector/internal/metric"
	"golang.org/x/sync/errgroup"

	log "github.com/sirupsen/logrus"
)

var (
	ErrStorageIsNil     = errors.New("metrics buf is nil")
	ErrClientIsNil      = errors.New("HTTP client is nil")
	ErrUnexpectedStatus = errors.New("unexpected response status")
)

type Agent struct {
	pollInterval   uint
	reportInterval uint
	sender         *Sender
	l              *log.Logger
}

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
		pollInterval:   cfg.PollInterval,
		reportInterval: cfg.ReportInterval,
		sender:         sender,
		l:              lg,
	}, nil
}

func (a *Agent) Run() error {
	a.l.Info("Agent is running...")

	var (
		scanCount int64
		storage   []metric.Metrics
		mtx       sync.Mutex
		err       error
	)

	scanTicker := time.NewTicker(time.Duration(a.pollInterval) * time.Second)
	sendTicker := time.NewTicker(time.Duration(a.reportInterval) * time.Second)

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil

			case <-scanTicker.C:
				metricCh := scanMetrics(ctx)

				mtx.Lock()

				storage = make([]metric.Metrics, 0)

				for mtrc := range metricCh {
					if mtrc.err != nil {
						return err
					}

					storage = append(storage, *mtrc.mtrc)
				}

				scanCount++

				mtx.Unlock()
			}
		}
	})

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil

			case <-sendTicker.C:
				mtx.Lock()

				pollCount, err := metric.NewMetrics("PollCount", "", scanCount)
				if err != nil {
					return err
				}

				storage = append(storage, *pollCount)
				scanCount = 0
				sndStorage := make([]metric.Metrics, len(storage))

				copy(sndStorage, storage)

				mtx.Unlock()

				if err = a.sender.Send(ctx, sndStorage); err != nil {
					return err
				}
			}
		}
	})

	return g.Wait()
}
