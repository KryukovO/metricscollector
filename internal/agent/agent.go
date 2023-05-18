package agent

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/KryukovO/metricscollector/internal/agent/config"
	"github.com/KryukovO/metricscollector/internal/metric"
)

var (
	ErrStorageIsNil = errors.New("metrics storage is nil")
	ErrClientIsNil  = errors.New("HTTP client is nil")
)

func Run(c *config.Config) error {
	log.Println("Agent is running...")

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
		if time.Since(lastReport) > time.Duration(c.ReportInterval)*time.Second {
			for mname, mval := range m {
				err := sendMetric(&client, c.ServerAddress, mname, mval)
				if err == ErrClientIsNil {
					return err
				}
				if err != nil {
					log.Printf("error sending '%s' metric value: %s\n", mname, err.Error())
				}
			}

			log.Println("metrics sent")
			lastReport = time.Now()
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

	var rtm runtime.MemStats

	runtime.ReadMemStats(&rtm)

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

func sendMetric(client *http.Client, sAddr, mname string, mval interface{}) error {
	if client == nil {
		return ErrClientIsNil
	}

	var url string
	if mname == "PollCount" {
		url = fmt.Sprintf("http://%s/update/%s/%s/%d", sAddr, metric.CounterMetric, mname, mval.(int64))
	} else {
		url = fmt.Sprintf("http://%s/update/%s/%s/%f", sAddr, metric.GaugeMetric, mname, mval.(float64))
	}

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "text/plain")

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
