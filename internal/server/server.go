package server

import (
	"log"

	"github.com/KryukovO/metricscollector/internal/server/config"
	"github.com/KryukovO/metricscollector/internal/server/handlers"
	"github.com/KryukovO/metricscollector/internal/storage/memstorage"
	"github.com/labstack/echo"
)

func Run(c *config.Config) error {
	// Инициализация хранилища
	s := memstorage.New()

	// Инициализация сервера
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	if err := handlers.SetHandlers(e, s); err != nil {
		return err
	}

	// Запуск сервера
	log.Printf("Server is running on %s...\n", c.HTTPAddress)
	if err := e.Start(c.HTTPAddress); err != nil {
		return err
	}

	return nil
}
