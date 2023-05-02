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
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func Run(c *config.Config) error {
	log.Println("Agent is running...")

	client := http.Client{
		//TODO: timeout?
	}

	m := make(map[string]interface{})
	var lastReport time.Time

	for {
		// сканируем метрики
		err := scanMetrics(m)
		if err != nil {
			return err
		}

		// отправляем метрики на сервер, если прошло reportInterval времени
		if time.Since(lastReport) > time.Duration(c.ReportInterval)*time.Second {
			err := sendMetrics(&client, m)
			if err != nil {
				return err
			}
			lastReport = time.Now()
		}

		// ожидание следующего интервала сканирования метрик
		time.Sleep(time.Duration(c.PollInterval) * time.Second)
	}
}

func scanMetrics(m map[string]interface{}) error {
	if m == nil {
		return errors.New("metrics storage is nil")
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

	m["RandomValue"] = float64(rnd.Intn(10000000))
	m["PollCount"] = m["PollCount"].(int64) + 1

	return nil
}

func sendMetrics(client *http.Client, m map[string]interface{}) error {
	if client == nil {
		return errors.New("HTTP client is nil")
	}

	for mname, val := range m {
		var url string
		if mname == "PollCount" {
			url = fmt.Sprintf("http://localhost:8080/update/%s/%s/%d", metric.CounterMetric, mname, val.(int64))
		} else {
			url = fmt.Sprintf("http://localhost:8080/update/%s/%s/%f", metric.GaugeMetric, mname, val.(float64))
		}

		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			log.Printf("error sending '%s' metric value: %s\n", mname, err.Error())
			continue
		}

		req.Header.Add("Content-Type", "text/plain")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("error sending '%s' metric value: %s\n", mname, err.Error())
			continue
		}

		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("error sending '%s' metric value: %s\n", mname, resp.Status)
		}
	}

	log.Println("metrics sent")
	return nil
}
