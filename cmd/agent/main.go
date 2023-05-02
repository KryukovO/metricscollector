package main

import (
	"flag"
	"log"

	"github.com/KryukovO/metricscollector/internal/agent/agent"
	"github.com/KryukovO/metricscollector/internal/agent/config"
)

func main() {
	c := config.New()

	flag.StringVar(&c.ServerAddress, "a", "localhost:8080", "Server endpoint address")
	flag.UintVar(&c.ReportInterval, "r", 10, "Metric reporting frequency in second")
	flag.UintVar(&c.PollInterval, "p", 2, "Metric polling frequency in seconds")
	flag.Parse()

	if err := agent.Run(c); err != nil {
		log.Fatalf("agent error: %s. Exit(1)\n", err.Error())
	}
}
