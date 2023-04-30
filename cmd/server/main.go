package main

import (
	"github.com/KryukovO/metricscollector/internal/server"
	"github.com/KryukovO/metricscollector/internal/storage/memstorage"
)

func main() {
	server.Run(memstorage.New())
}
