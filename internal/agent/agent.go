package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/KryukovO/metricscollector/internal/agent/config"
	"github.com/KryukovO/metricscollector/internal/metric"

	log "github.com/sirupsen/logrus"
)

var (
	ErrStorageIsNil = errors.New("metrics storage is nil")
	ErrClientIsNil  = errors.New("HTTP client is nil")
)

const BatchSize = 20 // ограничение количества метрик, отправляемых одним запросом

func Run(c *config.Config, l *log.Logger) error {
	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	lg.Info("Agent is running...")

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	client := http.Client{
		//TODO: timeout?
	}

	m := make(map[string]interface{})
	var lastReport time.Time
	var lastScan time.Time

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
			mtrcs := make([]metric.Metrics, 0, BatchSize)
			for mname, mval := range m {
				mtrc, err := metric.NewMetrics(mname, "", mval)
				if err != nil {
					lg.Infof("error sending '%s' metric value: %s", mname, err.Error())
					continue
				}
				mtrcs = append(mtrcs, *mtrc)

				if len(mtrcs) == BatchSize {
					err := sendMetrics(&client, c.ServerAddress, mtrcs)
					if err != nil {
						lg.Infof("error sending metric values: %s", err.Error())
					} else {
						lg.Infof("metrics sent: %d", len(mtrcs))
					}
					mtrcs = make([]metric.Metrics, 0, BatchSize)
				}
			}
			err := sendMetrics(&client, c.ServerAddress, mtrcs)
			if err != nil {
				lg.Infof("error sending metric values: %s", err.Error())
			} else {
				lg.Infof("metrics sent: %d", len(mtrcs))
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
// Если rnd == nil, то используется стандартный генератор math/rand
func scanMetrics(m map[string]interface{}, rnd *rand.Rand) error {
	if m == nil {
		return ErrStorageIsNil
	}

	if _, ok := m["PollCount"]; !ok {
		m["PollCount"] = int64(0)
	}

	rtm := &runtime.MemStats{}

	runtime.ReadMemStats(rtm)

	m["Alloc"] = float64(rtm.Alloc)
	m["BuckHashSys"] = float64(rtm.BuckHashSys)
	m["Frees"] = float64(rtm.Frees)
	m["GCCPUFraction"] = float64(rtm.GCCPUFraction)
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

	rndVal := float64(rand.Intn(10000000))
	if rnd != nil {
		rndVal = float64(rnd.Intn(10000000))
	}
	m["RandomValue"] = rndVal
	m["PollCount"] = m["PollCount"].(int64) + 1

	return nil
}

func sendMetrics(client *http.Client, sAddr string, mtrcs []metric.Metrics) error {
	if client == nil {
		return ErrClientIsNil
	}
	if len(mtrcs) == 0 {
		return nil
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

	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	io.Copy(io.Discard, resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}
