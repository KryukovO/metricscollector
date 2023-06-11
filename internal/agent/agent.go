package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"syscall"
	"time"

	"github.com/KryukovO/metricscollector/internal/agent/config"
	"github.com/KryukovO/metricscollector/internal/metric"

	log "github.com/sirupsen/logrus"
)

var (
	ErrStorageIsNil     = errors.New("metrics storage is nil")
	ErrClientIsNil      = errors.New("HTTP client is nil")
	ErrUnexpectedStatus = errors.New("unexpected response status")
)

const batchSize = 20 // ограничение количества метрик, отправляемых одним запросом

func Run(c *config.Config, l *log.Logger) error {
	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	lg.Info("Agent is running...")

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	client := http.Client{
		// NOTE: timeout?
	}

	var (
		m          = make(map[string]interface{})
		lastReport time.Time
		lastScan   time.Time
	)

	for {
		// сканируем метрики, если прошло pollInterval секунд с последнего сканирования
		if time.Since(lastScan) > time.Duration(c.PollInterval)*time.Second {
			err := scanMetrics(m, rnd)
			if err != nil {
				return err
			}

			lastScan = time.Now()
		}

		// отправляем метрики на сервер, если прошло reportInterval секунд с последней отправки
		// после отправки сбрасываем текущие сохраненные значения метрик
		if time.Since(lastReport) > time.Duration(c.ReportInterval)*time.Second {
			err := initMetricSend(&client, c.ServerAddress, m, lg)
			if err != nil {
				return err
			}

			lastReport = time.Now()
			m = make(map[string]interface{})
		}

		// выполняем проверку необходимости сканирования/отправки раз в секунду
		time.Sleep(time.Second)
	}
}

// Сканирование метрик в хранилище m.
//
// rnd - опциональный параметр, используемый для генерации случайной метрики RandomValue.
// Если rnd == nil, то используется стандартный генератор math/rand.
func scanMetrics(m map[string]interface{}, rnd *rand.Rand) error {
	if m == nil {
		return ErrStorageIsNil
	}

	var pollCount int64

	if val, ok := m["PollCount"]; !ok {
		m["PollCount"] = int64(0)
	} else {
		if v, ok := val.(int64); ok {
			pollCount = v
		}
	}

	rtm := &runtime.MemStats{}

	runtime.ReadMemStats(rtm)

	m["Alloc"] = float64(rtm.Alloc)
	m["BuckHashSys"] = float64(rtm.BuckHashSys)
	m["Frees"] = float64(rtm.Frees)
	m["GCCPUFraction"] = rtm.GCCPUFraction
	m["GCSys"] = float64(rtm.GCSys)
	m["HeapAlloc"] = float64(rtm.HeapAlloc)
	m["HeapIdle"] = float64(rtm.HeapIdle)
	m["HeapInuse"] = float64(rtm.HeapInuse)
	m["HeapObjects"] = float64(rtm.HeapObjects)
	m["HeapReleased"] = float64(rtm.HeapReleased)
	m["HeapSys"] = float64(rtm.HeapSys)
	m["LastGC"] = float64(rtm.LastGC)
	m["Lookups"] = float64(rtm.Lookups)
	m["MCacheInuse"] = float64(rtm.MCacheInuse)
	m["MCacheSys"] = float64(rtm.MCacheSys)
	m["MSpanInuse"] = float64(rtm.MSpanInuse)
	m["MSpanSys"] = float64(rtm.MSpanSys)
	m["Mallocs"] = float64(rtm.Mallocs)
	m["NextGC"] = float64(rtm.NextGC)
	m["NumForcedGC"] = float64(rtm.NumForcedGC)
	m["NumGC"] = float64(rtm.NumGC)
	m["OtherSys"] = float64(rtm.OtherSys)
	m["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	m["StackInuse"] = float64(rtm.StackInuse)
	m["StackSys"] = float64(rtm.StackSys)
	m["Sys"] = float64(rtm.Sys)
	m["TotalAlloc"] = float64(rtm.TotalAlloc)

	rndVal := float64(rand.Int())
	if rnd != nil {
		rndVal = float64(rnd.Int())
	}

	m["RandomValue"] = rndVal
	m["PollCount"] = pollCount + 1

	return nil
}

func initMetricSend(client *http.Client, address string, m map[string]interface{}, l *log.Logger) error {
	if m == nil {
		return ErrStorageIsNil
	}

	mtrcs := make([]metric.Metrics, 0, batchSize)

	send := func([]metric.Metrics) {
		var err error

		for t := 1; t <= 5; t += 2 {
			err = sendMetrics(client, address, mtrcs)
			if err == nil || !errors.Is(err, syscall.ECONNREFUSED) {
				break
			}

			time.Sleep(time.Duration(t) * time.Second)
		}

		if err != nil {
			l.Infof("error sending metric values: %s", err.Error())
		} else {
			l.Infof("metrics sent: %d", len(mtrcs))
		}
	}

	for mname, mval := range m {
		mtrc, err := metric.NewMetrics(mname, "", mval)
		if err != nil {
			l.Infof("error sending '%s' metric value: %s", mname, err.Error())

			continue
		}

		mtrcs = append(mtrcs, *mtrc)

		if len(mtrcs) == batchSize {
			send(mtrcs)
			mtrcs = make([]metric.Metrics, 0, batchSize)
		}
	}

	if len(mtrcs) > 0 {
		send(mtrcs)
	}

	return nil
}

func sendMetrics(client *http.Client, sAddr string, mtrcs []metric.Metrics) error {
	if client == nil {
		return ErrClientIsNil
	}

	url := fmt.Sprintf("http://%s/updates/", sAddr)

	body, err := json.Marshal(mtrcs)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}

	gz := gzip.NewWriter(buf)
	if _, err = gz.Write(body); err != nil {
		return err
	}

	gz.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, buf)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	_, err = io.Copy(io.Discard, resp.Body)

	defer resp.Body.Close()

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %w", resp.Status, ErrUnexpectedStatus)
	}

	return nil
}
