package server

import (
	"database/sql"
	"time"

	"github.com/KryukovO/metricscollector/internal/server/config"
	"github.com/KryukovO/metricscollector/internal/server/handlers"
	"github.com/KryukovO/metricscollector/internal/storage/memstorage"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

func Run(c *config.Config, l *log.Logger) error {
	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	// Инициализация хранилища
	s, err := memstorage.NewMemStorage(c.FileStoragePath, c.Restore, time.Duration(c.StoreInterval)*time.Second)
	if err != nil {
		return err
	}
	defer s.Close()

	// Подключение к БД
	db, err := sql.Open("pgx", c.DSN)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err == nil {
		lg.Info("Database connection established")
		defer func() {
			db.Close()
			lg.Info("Database connection closed")
		}()
	}

	// Инициализация сервера
	// TODO: переопределить e.HTTPErrorHandler, чтобы он не заполнял тело ответа
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	if err := handlers.SetHandlers(e, s, lg, db); err != nil {
		return err
	}

	// Запуск сервера
	lg.Infof("Server is running on %s...", c.HTTPAddress)
	if err := e.Start(c.HTTPAddress); err != nil {
		return err
	}

	return nil
}
