package main

import (
	"flag"

	"github.com/KryukovO/metricscollector/internal/agent"
	"github.com/KryukovO/metricscollector/internal/agent/config"

	"github.com/caarlos0/env"
	log "github.com/sirupsen/logrus"
)

func main() {
	c := config.NewConfig()

	flag.StringVar(&c.ServerAddress, "a", "localhost:8080", "Server endpoint address")
	flag.UintVar(&c.ReportInterval, "r", 10, "Metric reporting frequency in second")
	flag.UintVar(&c.PollInterval, "p", 2, "Metric polling frequency in seconds")
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05 Z07:00",
	})

	err := env.Parse(c)
	if err != nil {
		log.Fatalf("env parsing error: %s. Exit(1)", err.Error())
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05 Z07:00",
	})

	if err := agent.Run(c); err != nil {
		log.Fatalf("agent error: %s. Exit(1)", err.Error())
	}
}
