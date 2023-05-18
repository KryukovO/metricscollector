package main

import (
	"flag"
	"log"

	"github.com/KryukovO/metricscollector/internal/server"
	"github.com/KryukovO/metricscollector/internal/server/config"
	"github.com/caarlos0/env"
)

func main() {
	c := config.NewConfig()

	flag.StringVar(&c.HTTPAddress, "a", "localhost:8080", "Server endpoint address")
	flag.Parse()

	err := env.Parse(c)
	if err != nil {
		log.Fatalf("env parsing error: %s. Exit(1)\n", err.Error())
	}

	if err := server.Run(c); err != nil {
		log.Fatalf("server error: %s. Exit(1)\n", err.Error())
	}
}
