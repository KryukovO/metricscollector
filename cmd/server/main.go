package main

import (
	"flag"
	"log"

	"github.com/KryukovO/metricscollector/internal/server/config"
	"github.com/KryukovO/metricscollector/internal/server/server"
)

func main() {
	c := config.New()

	flag.StringVar(&c.HTTPAddress, "a", "127.0.0.1:8080", "Server endpoint address")
	flag.Parse()

	if err := server.Run(c); err != nil {
		log.Fatalf("server error: %s. Exit(1)\n", err.Error())
	}
}
