package middleware

import (
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

func LoggingMiddlewarefunc(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(e echo.Context) (err error) {
		log.Infof("recieved query with method %s: %s", e.Request().Method, e.Request().RequestURI)
		return next(e)
	})
}
