// Модуль-агент предназначен для сбора метрик из различных источников
// и передачи их в хранилище.
package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"

	"github.com/KryukovO/metricscollector/internal/agent"
	"github.com/KryukovO/metricscollector/internal/agent/config"

	"github.com/caarlos0/env"
	log "github.com/sirupsen/logrus"

	_ "net/http/pprof"
)

var (
	// buildVersion представляет собой хранилище значения ldflag - версия сборки.
	buildVersion = "N/A"
	// buildDate представляет собой хранилище значения ldflag - дата сборки.
	buildDate = "N/A"
	// buildCommit представляет собой хранилище значения ldflag - комментарий к сборке.
	buildCommit = "N/A"
)

const (
	serverAddress  = "localhost:8080" // Адрес сервера-хранилища по умолчанию
	reportInterval = 10               // Интервал отправки метрик в хранилище в секундах по умолчанию
	pollInterval   = 2                // Интервал сканирования метрик в секундах по умолчанию
	key            = ""               // Значения ключа аутентификации по умолчанию
	rateLimit      = 2                // Количество одновременно исходящих запросов на сервер по умолчанию

	httpTimeout = 5       // Таймаут соединения с сервером по умолчанию
	batchSize   = 10      // Количество посылаемых за раз метрик по умолчанию
	retries     = "1,3,5" // Интервалы попыток соединения с сервером через запятую по умолчанию
)

func main() {
	fmt.Printf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		buildVersion, buildDate, buildCommit,
	)

	cfg := config.NewConfig()

	flag.StringVar(&cfg.ServerAddress, "a", serverAddress, "Server endpoint address")
	flag.UintVar(&cfg.ReportInterval, "r", reportInterval, "Metric reporting frequency in second")
	flag.UintVar(&cfg.PollInterval, "p", pollInterval, "Metric polling frequency in seconds")
	flag.StringVar(&cfg.Key, "k", key, "Server key")
	flag.UintVar(&cfg.RateLimit, "l", rateLimit, "Number of concurrent requests")

	flag.UintVar(&cfg.HTTPTimeout, "timeout", httpTimeout, "Server connection timeout")
	flag.UintVar(&cfg.BatchSize, "batch", batchSize, "Metrics batch size")
	flag.StringVar(&cfg.Retries, "retries", retries, "Server connection retry intervals")

	flag.Parse()

	l := log.New()
	l.SetLevel(log.DebugLevel)
	l.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05 Z07:00",
	})

	go func() {
		if err := http.ListenAndServe("0.0.0.0:8081", nil); !errors.Is(err, http.ErrServerClosed) {
			l.Errorf("pprof running error: %v", err)
		}
	}()

	err := env.Parse(cfg)
	if err != nil {
		l.Fatalf("env parsing error: %s. Exit(1)", err.Error())
	}

	agnt, err := agent.NewAgent(cfg, l)
	if err != nil {
		l.Fatalf("agent initialization error: %s. Exit(1)", err.Error())
	}

	if err := agnt.Run(); err != nil {
		l.Fatalf("agent running error: %s. Exit(1)", err.Error())
	}
}
