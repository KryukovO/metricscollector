// Модуль-сервер предназначен для приема метрик от агента
// и сохранения их в различные хранилища.
package main

import (
	"context"
	"flag"

	"github.com/KryukovO/metricscollector/internal/server"
	"github.com/KryukovO/metricscollector/internal/server/config"

	"github.com/caarlos0/env"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	log "github.com/sirupsen/logrus"
)

const (
	httpAddress     = "localhost:8080"       // Адрес эндпоинта сервера (host:port) по умолчанию
	storeInterval   = 300                    // Интервал сохранения значения метрик в файл в секундах по умолчанию
	fileStoragePath = "/tmp/metrics-db.json" // Полное имя файла, куда сохраняются текущие значения метрик по умолчанию
	restore         = true                   // Признак загрузки значений метрик из файла при запуске сервера по умолчанию
	dsn             = ""                     // Адрес подключения к БД по умолчанию
	key             = ""                     // Ключ аутентификации по умолчанию

	storeTimeout    = 5                // Таймаут выполнения операций с хранилищем по умолчанию
	shutdownTimeout = 10               // Таймаут для graceful shutdown сервера по умолчанию
	retries         = "1,3,5"          // Интервалы попыток соединения с хранилищем через запятую по умолчанию
	migrations      = "sql/migrations" // Путь до директории с файлами миграции по умолчанию
)

func main() {
	cfg := config.NewConfig()

	flag.StringVar(&cfg.HTTPAddress, "a", httpAddress, "Server endpoint address")
	flag.UintVar(&cfg.StoreInterval, "i", storeInterval, "Store interval")
	flag.StringVar(&cfg.FileStoragePath, "f", fileStoragePath, "File storage path")
	flag.BoolVar(&cfg.Restore, "r", restore, "Restore")
	flag.StringVar(&cfg.DSN, "d", dsn, "Data source name")
	flag.StringVar(&cfg.Key, "k", key, "Server key")

	flag.UintVar(&cfg.StoreTimeout, "timeout", storeTimeout, "Storage connection timeout")
	flag.UintVar(&cfg.ShutdownTimeout, "shutdown", shutdownTimeout, "Graceful shutdown timeout")
	flag.StringVar(&cfg.Retries, "retries", retries, "Server connect retry intervals")
	flag.StringVar(&cfg.Migrations, "migrations", migrations, "Directory of database migration files")

	flag.Parse()

	l := log.New()
	l.SetLevel(log.DebugLevel)
	l.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05 Z07:00",
	})

	err := env.Parse(cfg)
	if err != nil {
		l.Fatalf("Env parsing error: %s. Exit(1)", err.Error())
	}

	s := server.NewServer(cfg, l)
	if err := s.Run(context.Background()); err != nil {
		l.Fatalf("Server error: %s. Exit(1)", err.Error())
	}
}
