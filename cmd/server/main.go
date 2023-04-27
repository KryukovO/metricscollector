package main

import (
	"github.com/KryukovO/metricscollector/internal/server"
	"github.com/KryukovO/metricscollector/internal/storage/memstorage"
)

func main() {
	storage := memstorage.New()
	server := server.New(storage)
	server.Run()
}
