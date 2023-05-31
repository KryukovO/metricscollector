package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/KryukovO/metricscollector/internal/server/middleware/models"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

type MiddlewareManager struct {
	l *log.Logger
}

func NewMiddlewareManager(l *log.Logger) *MiddlewareManager {
	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}
	return &MiddlewareManager{l: lg}
}

func (mw *MiddlewareManager) LoggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(e echo.Context) (err error) {
		mw.l.Infof("recieved query with method %s: %s", e.Request().Method, e.Request().RequestURI)

		ts := time.Now()
		res := next(e)

		if res != nil {
			mw.l.Infof("query response status: %d; size: %d; duration: %s", res.(*echo.HTTPError).Code, e.Response().Size, time.Since(ts))
		} else {
			mw.l.Infof("query response status: %d; size: %d; duration: %s", e.Response().Status, e.Response().Size, time.Since(ts))
		}

		return res
	})
}

func (mw *MiddlewareManager) GZipMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(e echo.Context) (err error) {
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
