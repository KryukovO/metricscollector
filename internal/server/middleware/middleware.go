package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/KryukovO/metricscollector/internal/server/middleware/models"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

type Manager struct {
	l *log.Logger
}

func NewManager(l *log.Logger) *Manager {
	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	return &Manager{l: lg}
}

func (mw *Manager) LoggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(e echo.Context) error {
		mw.l.Infof("received query with method %s: %s", e.Request().Method, e.Request().RequestURI)

		ts := time.Now()
		err := next(e)

		if err != nil {
			var echoErr *echo.HTTPError
			if errors.As(err, &echoErr) {
				mw.l.Infof(
					"query response status: %d; size: %d; duration: %s",
					echoErr.Code, e.Response().Size, time.Since(ts),
				)
			}
		} else {
			mw.l.Infof(
				"query response status: %d; size: %d; duration: %s",
				e.Response().Status, e.Response().Size, time.Since(ts),
			)
		}

		return err
	})
}

func (mw *Manager) GZipMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(e echo.Context) error {
		// Допустимые для сжатия форматы данных в ответе
		acceptTypes := []string{"application/json", "text/html"}

		// Проверка допускает ли клиент данные сжатые gzip
		acceptEnc := e.Request().Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEnc, "gzip")

		if supportsGzip {
			cw := models.NewCompressWriter(e.Response().Writer, acceptTypes)
			e.Response().Writer = cw
		}

		// Проверка сжаты ли данные запроса
		contentEncoding := e.Request().Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := models.NewCompressReader(e.Request().Body)
			if err != nil {
				mw.l.Infof("something went wrong: %s", err.Error())

				return e.NoContent(http.StatusInternalServerError)
			}

			// Меняем тело запроса на новое
			e.Request().Body = cr
			defer cr.Close()
		}

		return next(e)
	})
}
