package middleware

import (
	"strings"
	"time"

	"github.com/KryukovO/metricscollector/internal/server/middleware/models"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

func LoggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(e echo.Context) (err error) {
		log.Infof("recieved query with method %s: %s", e.Request().Method, e.Request().RequestURI)

		ts := time.Now()
		res := next(e)

		if res != nil {
			log.Infof("query response status: %d; size: %d; duration: %s", res.(*echo.HTTPError).Code, e.Response().Size, time.Since(ts))
		} else {
			log.Infof("query response status: %d; size: %d; duration: %s", e.Response().Status, e.Response().Size, time.Since(ts))
		}

		return res
	})
}

func GZipMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(e echo.Context) (err error) {
		// Допустимые для сжатия форматы данных в ответе
		acceptTypes := []string{"application/json", "text/html"}

		// Проверка допускает ли клиент данные сжатые gzip
		acceptEnc := e.Request().Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEnc, "gzip")

		if supportsGzip {
			cw := models.NewCompressWriter(e.Response().Writer, acceptTypes)
			e.Response().Writer = cw
			// Note: Закрытие cw вызывается в методе cw.Write
		}

		return next(e)
	})
}
