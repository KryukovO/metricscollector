package handlers

import (
	"errors"

	"github.com/KryukovO/metricscollector/internal/server/middleware"
	"github.com/KryukovO/metricscollector/internal/storage"
	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

var (
	ErrServerIsNil  = errors.New("server instance is nil")
	ErrStorageIsNil = errors.New("storage is nil")
)

func SetHandlers(e *echo.Echo, s storage.Storage, key []byte, l *log.Logger) error {
	if e == nil {
		return ErrServerIsNil
	}

	if s == nil {
		return ErrStorageIsNil
	}

	ctrl, err := NewStorageController(s, l)
	if err != nil {
		return err
	}

	mw := middleware.NewManager(key, l)
	e.Use(
		mw.LoggingMiddleware,
		mw.GZipMiddleware,
		mw.HashMiddleware,
	)

	pprof.Register(e) // Регистрация профилировщика

	return MapStorageHandlers(e.Router(), ctrl)
}
