package handlers

import (
	"errors"

	"github.com/KryukovO/metricscollector/internal/server/middleware"
	"github.com/KryukovO/metricscollector/internal/storage"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

var ErrServerIsNil = errors.New("server instance is nil")

func SetHandlers(e *echo.Echo, s storage.Storage, l *log.Logger) error {
	if e == nil {
		return ErrServerIsNil
	}

	if err := setStorageHandlers(e.Router(), s, l); err != nil {
		return err
	}

	mw := middleware.NewManager(l)
	e.Use(
		mw.LoggingMiddleware,
		mw.GZipMiddleware,
	)

	return nil
}
