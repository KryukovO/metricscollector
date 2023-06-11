package server

import (
	"time"

	"github.com/KryukovO/metricscollector/internal/server/config"
	"github.com/KryukovO/metricscollector/internal/server/handlers"
	"github.com/KryukovO/metricscollector/internal/storage"
	"github.com/KryukovO/metricscollector/internal/storage/repository/memstorage"
	"github.com/KryukovO/metricscollector/internal/storage/repository/pgstorage"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

func Run(c *config.Config, l *log.Logger) error {
	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	// Инициализация хранилища
	var (
		repo storage.Repo
		err  error
	)

	if c.DSN != "" {
		repo, err = pgstorage.NewPgStorage(c.DSN)
	} else {
		repo, err = memstorage.NewMemStorage(c.FileStoragePath, c.Restore, time.Duration(c.StoreInterval)*time.Second, lg)
	}

	if err != nil {
		return err
	}

	s := storage.NewMetricsStorage(repo)
	defer s.Close()

	// Инициализация сервера
	// NOTE: можно также переопределить e.HTTPErrorHandler, чтобы он не заполнял тело ответа
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	if err := handlers.SetHandlers(e, s, lg); err != nil {
		return err
	}

	// Запуск сервера
	lg.Infof("Server is running on %s...", c.HTTPAddress)

	return e.Start(c.HTTPAddress)
}
