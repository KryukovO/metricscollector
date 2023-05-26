package server

import (
	"github.com/KryukovO/metricscollector/internal/server/config"
	"github.com/KryukovO/metricscollector/internal/server/handlers"
	"github.com/KryukovO/metricscollector/internal/storage/memstorage"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

func Run(c *config.Config) error {
	// Инициализация хранилища
	s := memstorage.NewMemStorage()

	// Инициализация сервера
	// TODO: переопределить e.HTTPErrorHandler, чтобы он не заполнял тело ответа
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	if err := handlers.SetHandlers(e, s); err != nil {
		return err
	}

	// Запуск сервера
	log.Infof("Server is running on %s...", c.HTTPAddress)
	if err := e.Start(c.HTTPAddress); err != nil {
		return err
	}

	return nil
}
