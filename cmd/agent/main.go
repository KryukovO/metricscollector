package main

import (
	"log"

	"github.com/KryukovO/metricscollector/internal/agent"
)

func main() {
	err := agent.Run()
	log.Fatalf("error during agent operation: %s. Exit(1)", err.Error())
}
