package handlers

import (
	"errors"

	"github.com/KryukovO/metricscollector/internal/server/middleware"
	"github.com/KryukovO/metricscollector/internal/storage"
	"github.com/labstack/echo"
)

func SetHandlers(e *echo.Echo, s storage.Storage) error {
	if e == nil {
		return errors.New("server instance is nil")
	}

	if err := newStorageHandlers(e.Router(), s); err != nil {
		return err
	}

	e.Use(middleware.LoggingMiddlewarefunc)

	return nil
}
