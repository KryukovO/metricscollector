package main

import (
	"log"

	"github.com/KryukovO/metricscollector/internal/server"
	"github.com/KryukovO/metricscollector/internal/storage/memstorage"
)

func main() {
	if err := server.Run(memstorage.New()); err != nil {
		log.Fatalf("server error: %s. Exit(1)\n", err.Error())
	}
}
