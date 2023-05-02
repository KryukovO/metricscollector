package server

import (
	"log"

	"github.com/KryukovO/metricscollector/internal/server/handlers"
	"github.com/KryukovO/metricscollector/internal/storage"
	"github.com/labstack/echo"
)

func Run(s storage.Storage) error {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	if err := handlers.SetHandlers(e, s); err != nil {
		return err
	}

	log.Println("Server is running...")
	if err := e.Start(":8080"); err != nil {
		return err
	}

	return nil
}
