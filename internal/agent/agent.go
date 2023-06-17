package agent

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/KryukovO/metricscollector/internal/agent/config"
	"github.com/KryukovO/metricscollector/internal/metric"

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

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	var (
		scanCount  int64
		storage    []metric.Metrics
		lastReport time.Time
		lastScan   time.Time
		err        error
	)

	for {
		// сканируем метрики, если прошло pollInterval секунд с последнего сканирования
		if time.Since(lastScan) > time.Duration(a.pollInterval)*time.Second {
			storage, err = scanMetrics(rnd)
			if err != nil {
				return err
			}

			lastScan = time.Now()
			scanCount++
		}

		// отправляем метрики на сервер, если прошло reportInterval секунд с последней отправки
		// после отправки сбрасываем счётчик сканирований
		if time.Since(lastReport) > time.Duration(a.reportInterval)*time.Second {
			pollCount, err := metric.NewMetrics("PollCount", "", scanCount)
			if err != nil {
				return err
			}

			storage = append(storage, *pollCount)

			err = a.sender.InitMetricSend(storage)
			if err != nil {
				return err
			}

			lastReport = time.Now()
			scanCount = 0
		}

		// выполняем проверку необходимости сканирования/отправки раз в секунду
		time.Sleep(time.Second)
	}
}

// Сканирование метрик в хранилище.
//
// rnd - опциональный параметр, используемый для генерации случайной метрики RandomValue.
// Если rnd == nil, то используется стандартный генератор math/rand.
func scanMetrics(rnd *rand.Rand) ([]metric.Metrics, error) {
	buf := make(map[string]interface{})

	rndVal := float64(rand.Int())
	if rnd != nil {
		rndVal = float64(rnd.Int())
	}

	buf["RandomValue"] = rndVal

	rtm := &runtime.MemStats{}

	runtime.ReadMemStats(rtm)

	buf["Alloc"] = float64(rtm.Alloc)
	buf["BuckHashSys"] = float64(rtm.BuckHashSys)
	buf["Frees"] = float64(rtm.Frees)
	buf["GCCPUFraction"] = rtm.GCCPUFraction
	buf["GCSys"] = float64(rtm.GCSys)
	buf["HeapAlloc"] = float64(rtm.HeapAlloc)
	buf["HeapIdle"] = float64(rtm.HeapIdle)
	buf["HeapInuse"] = float64(rtm.HeapInuse)
	buf["HeapObjects"] = float64(rtm.HeapObjects)
	buf["HeapReleased"] = float64(rtm.HeapReleased)
	buf["HeapSys"] = float64(rtm.HeapSys)
	buf["LastGC"] = float64(rtm.LastGC)
	buf["Lookups"] = float64(rtm.Lookups)
	buf["MCacheInuse"] = float64(rtm.MCacheInuse)
	buf["MCacheSys"] = float64(rtm.MCacheSys)
	buf["MSpanInuse"] = float64(rtm.MSpanInuse)
	buf["MSpanSys"] = float64(rtm.MSpanSys)
	buf["Mallocs"] = float64(rtm.Mallocs)
	buf["NextGC"] = float64(rtm.NextGC)
	buf["NumForcedGC"] = float64(rtm.NumForcedGC)
	buf["NumGC"] = float64(rtm.NumGC)
	buf["OtherSys"] = float64(rtm.OtherSys)
	buf["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	buf["StackInuse"] = float64(rtm.StackInuse)
	buf["StackSys"] = float64(rtm.StackSys)
	buf["Sys"] = float64(rtm.Sys)
	buf["TotalAlloc"] = float64(rtm.TotalAlloc)

	storage := make([]metric.Metrics, 0, len(buf)+1)

	for mName, mVal := range buf {
		mtrc, err := metric.NewMetrics(mName, "", mVal)
		if err != nil {
			return nil, err
		}

		storage = append(storage, *mtrc)
	}

	return storage, nil
}
