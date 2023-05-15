package main

import (
	"flag"
	"log"

	"github.com/KryukovO/metricscollector/internal/agent/agent"
	"github.com/KryukovO/metricscollector/internal/agent/config"
	"github.com/caarlos0/env"
)

func main() {
	c := config.NewConfig()

	flag.StringVar(&c.ServerAddress, "a", "localhost:8080", "Server endpoint address")
	flag.UintVar(&c.ReportInterval, "r", 10, "Metric reporting frequency in second")
	flag.UintVar(&c.PollInterval, "p", 2, "Metric polling frequency in seconds")
	flag.Parse()

	err := env.Parse(c)
	if err != nil {
		log.Fatalf("env parsing error: %s. Exit(1)\n", err.Error())
	}

	if err := agent.Run(c); err != nil {
		log.Fatalf("agent error: %s. Exit(1)\n", err.Error())
	}
}
