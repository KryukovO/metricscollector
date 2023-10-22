// Модуль-сервер предназначен для приема метрик от агента
// и сохранения их в различные хранилища.
package main

import (
	"context"
	"fmt"

	"github.com/KryukovO/metricscollector/internal/server"
	"github.com/KryukovO/metricscollector/internal/server/config"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	log "github.com/sirupsen/logrus"
)

var (
	// buildVersion представляет собой хранилище значения ldflag - версия сборки.
	buildVersion = "N/A"
	// buildDate представляет собой хранилище значения ldflag - дата сборки.
	buildDate = "N/A"
	// buildCommit представляет собой хранилище значения ldflag - комментарий к сборке.
	buildCommit = "N/A"
)

func main() {
	fmt.Printf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		buildVersion, buildDate, buildCommit,
	)

	l := log.New()
	l.SetLevel(log.DebugLevel)
	l.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05 Z07:00",
	})

	cfg, err := config.NewConfig()
	if err != nil {
		l.Fatalf("config init error: %v. Exit(1)", err)
	}

	s := server.NewServer(cfg, l)
	if err := s.Run(context.Background()); err != nil {
		l.Fatalf("Server error: %s. Exit(1)", err.Error())
	}
}
