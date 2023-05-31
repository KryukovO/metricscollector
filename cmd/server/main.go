package main

import (
	"flag"

	"github.com/KryukovO/metricscollector/internal/server"
	"github.com/KryukovO/metricscollector/internal/server/config"

	"github.com/caarlos0/env"
	_ "github.com/jackc/pgx/v5/stdlib"
	log "github.com/sirupsen/logrus"
)

func main() {
	c := config.NewConfig()

	flag.StringVar(&c.HTTPAddress, "a", "localhost:8080", "Server endpoint address")
	flag.UintVar(&c.StoreInterval, "i", 300, "Store interval")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/metrics-db.json", "File storage path")
	flag.BoolVar(&c.Restore, "r", true, "Restore")
	flag.StringVar(&c.DSN, "d", "", "Data source name")
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

	if err := server.Run(c, l); err != nil {
		l.Fatalf("server error: %s. Exit(1)", err.Error())
	}
}
