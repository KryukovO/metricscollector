package main

import (
	"log"

	"github.com/KryukovO/metricscollector/internal/agent"
)

func main() {
	if err := agent.Run(); err != nil {
		log.Fatalf("agent error: %s. Exit(1)\n", err.Error())
	}
}
