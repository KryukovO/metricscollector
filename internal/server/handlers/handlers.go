package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/KryukovO/metricscollector/internal/server/middleware"
	"github.com/KryukovO/metricscollector/internal/storage"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

var dbcon *sql.DB

func SetHandlers(e *echo.Echo, s storage.Storage, l *log.Logger, db *sql.DB) error {
	if e == nil {
		return errors.New("server instance is nil")
	}

	if err := setStorageHandlers(e.Router(), s, l); err != nil {
		return err
	}
	e.Router().Add(http.MethodGet, "/ping", pingHandler)

	mw := middleware.NewMiddlewareManager(l)
	e.Use(
		mw.LoggingMiddleware,
		mw.GZipMiddleware,
	)

	dbcon = db

	return nil
}

func pingHandler(e echo.Context) error {
	err := dbcon.Ping()
	if err != nil {
		return e.NoContent(http.StatusInternalServerError)
	}
	return e.NoContent(http.StatusOK)
}
