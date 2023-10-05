package handlers

import (
	"errors"

	"github.com/KryukovO/metricscollector/internal/server/middleware"
	"github.com/KryukovO/metricscollector/internal/storage"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

var (
	// Возвращается SetHandlers, если передан неициализированный инстанс echo.
	ErrServerIsNil = errors.New("server instance is nil")
	// Возвращается SetHandlers и NewStorageController, если передано неинициализированное хранилище.
	ErrStorageIsNil = errors.New("storage is nil")
)

// Инициирует маппинг маршрутов и обработчиков в инстанс echo,
// а также выстраивает цепочку middleware.
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

	return MapStorageHandlers(e.Router(), ctrl)
}
