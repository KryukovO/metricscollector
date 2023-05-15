package middleware

import (
	"log"

	"github.com/labstack/echo"
)

func LoggingMiddlewarefunc(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(e echo.Context) (err error) {
		log.Printf("recieved query with method %s: %s\n", e.Request().Method, e.Request().RequestURI)
		return next(e)
	})
}
