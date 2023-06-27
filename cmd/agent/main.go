package main

import (
	"flag"

	"github.com/KryukovO/metricscollector/internal/agent"
	"github.com/KryukovO/metricscollector/internal/agent/config"

	"github.com/caarlos0/env"
	log "github.com/sirupsen/logrus"
)

const (
	serverAddress  = "localhost:8080"
	reportInterval = 10
	pollInterval   = 2
	key            = ""
	rateLimit      = 2

	httpTimeout = 5
	batchSize   = 10
	retries     = "1,3,5"
)

func main() {
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
