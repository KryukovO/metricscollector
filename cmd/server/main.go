package main

import (
	"flag"

	"github.com/KryukovO/metricscollector/internal/server"
	"github.com/KryukovO/metricscollector/internal/server/config"

	"github.com/caarlos0/env"
	_ "github.com/jackc/pgx/v5/stdlib"
	log "github.com/sirupsen/logrus"
)

const (
	httpAddress     = "localhost:8080"
	storeInterval   = 300
	fileStoragePath = "/tmp/metrics-db.json"
	restore         = true
	dsn             = ""
	storageTimeout  = 5
	retries         = "1,3,5"
)

func main() {
	cfg := config.NewConfig()

	flag.StringVar(&cfg.HTTPAddress, "a", httpAddress, "Server endpoint address")
	flag.UintVar(&cfg.StoreInterval, "i", storeInterval, "Store interval")
	flag.StringVar(&cfg.FileStoragePath, "f", fileStoragePath, "File storage path")
	flag.BoolVar(&cfg.Restore, "r", restore, "Restore")
	flag.StringVar(&cfg.DSN, "d", dsn, "Data source name")
	flag.UintVar(&cfg.StorageTimeout, "timeout", storageTimeout, "Storage connection timeout")
	flag.StringVar(&cfg.Retries, "retries", retries, "Server connect retry intervals")
	flag.Parse()

	l := log.New()
	l.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05 Z07:00",
	})

	err := env.Parse(cfg)
	if err != nil {
		l.Fatalf("env parsing error: %s. Exit(1)", err.Error())
	}

	s := server.NewServer(cfg, l)
	if err := s.Run(); err != nil {
		l.Fatalf("server error: %s. Exit(1)", err.Error())
	}
}
