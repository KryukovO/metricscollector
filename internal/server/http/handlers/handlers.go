// Package handlers содержит обработчики http-запросов к серверу.
package handlers

import (
	"crypto/rsa"
	"errors"
	"net"

	"github.com/KryukovO/metricscollector/internal/server/http/middleware"
	"github.com/KryukovO/metricscollector/internal/storage"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrServerIsNil возвращается SetHandlers, если передан неициализированный инстанс echo.
	ErrServerIsNil = errors.New("server instance is nil")
	// ErrStorageIsNil возвращается SetHandlers и NewStorageController, если передано неинициализированное хранилище.
	ErrStorageIsNil = errors.New("storage is nil")
)

// SetHandlers инициирует маппинг маршрутов и обработчиков в инстанс echo,
// а также выстраивает цепочку middleware.
func SetHandlers(
	e *echo.Echo, s storage.Storage,
	key []byte, privateKey *rsa.PrivateKey, trustedSNet *net.IPNet,
	l *log.Logger,
) error {
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

	mw := middleware.NewManager(key, privateKey, trustedSNet, l)
	e.Use(
		mw.LoggingMiddleware,
		mw.IPValidationMiddleware,
		mw.GZipMiddleware,
		mw.HashMiddleware,
		mw.RSAMiddleware,
	)

	return MapStorageHandlers(e.Router(), ctrl)
}
