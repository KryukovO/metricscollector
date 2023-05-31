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

	l := log.New()
	l.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05 Z07:00",
	})

	err := env.Parse(c)
	if err != nil {
		l.Fatalf("env parsing error: %s. Exit(1)", err.Error())
	}

	if err := agent.Run(c, l); err != nil {
		l.Fatalf("agent error: %s. Exit(1)", err.Error())
	}
}
