package middleware

import (
	"time"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

func LoggingMiddlewarefunc(next echo.HandlerFunc) echo.HandlerFunc {
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
