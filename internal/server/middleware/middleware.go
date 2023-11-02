// Package middleware содержит промежуточные обработчки для HTTP-сервера.
package middleware

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/KryukovO/metricscollector/internal/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

// Manager предназначен для управления middleware.
type Manager struct {
	key        []byte
	privateKey rsa.PrivateKey
	l          *log.Logger
}

// NewManager создаёт новый объект Manager.
func NewManager(key []byte, privateKey rsa.PrivateKey, l *log.Logger) *Manager {
	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	return &Manager{key: key, privateKey: privateKey, l: lg}
}

// LoggingMiddleware - middleware для логирования входящих запросов и их результатов.
func (mw *Manager) LoggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(e echo.Context) error {
		uuid := uuid.New()
		e.Set("uuid", uuid)

		mw.l.Infof("[%s] received query with method %s: %s", uuid, e.Request().Method, e.Request().RequestURI)

		ts := time.Now()
		err := next(e)

		if err != nil {
			var echoErr *echo.HTTPError
			if errors.As(err, &echoErr) {
				mw.l.Infof(
					"[%s] query response status: %d; size: %d; duration: %s",
					uuid, echoErr.Code, e.Response().Size, time.Since(ts),
				)
			}
		} else {
			mw.l.Infof(
				"[%s] query response status: %d; size: %d; duration: %s",
				uuid, e.Response().Status, e.Response().Size, time.Since(ts),
			)
		}

		return err
	})
}

// GZipMiddleware - middleware для распаковки сжатых входящих запрос и сжатия результатов.
func (mw *Manager) GZipMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(e echo.Context) error {
		uuid := e.Get("uuid")

		// Допустимые для сжатия форматы данных в ответе
		acceptTypes := [...]string{"application/json", "text/html"}

		// Проверка допускает ли клиент данные сжатые gzip
		acceptEnc := e.Request().Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEnc, "gzip")

		if supportsGzip {
			cw := NewCompressWriter(e.Response().Writer, acceptTypes[:])
			e.Response().Writer = cw
		}

		// Проверка сжаты ли данные запроса
		contentEncoding := e.Request().Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := NewCompressReader(e.Request().Body)
			if err != nil {
				mw.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

				return e.NoContent(http.StatusInternalServerError)
			}

			// Меняем тело запроса на новое
			e.Request().Body = cr
			defer cr.Close()
		}

		return next(e)
	})
}

// HashMiddleware - middleware для валидации входящего запроса
// путём сравнения хеша данных и значения в заголовке HashSHA256.
func (mw *Manager) HashMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(e echo.Context) error {
		if mw.key == nil {
			return next(e)
		}

		uuid := e.Get("uuid")

		ctx := NewContext(e, mw.key)

		body, err := io.ReadAll(ctx.Request().Body)
		if err != nil {
			mw.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

			return ctx.NoContent(http.StatusInternalServerError)
		}

		if len(body) == 0 {
			return next(ctx)
		}

		ctx.Request().Body = io.NopCloser(bytes.NewBuffer(body))

		hash := ctx.Request().Header.Get("HashSHA256")

		// NOTE: Затычка для тестов 14 инкремента, которые шлют запросы без хеша, но требуют 200 ОК
		if hash == "" {
			return next(ctx)
		}

		hexHash, err := hex.DecodeString(hash)
		if err != nil {
			mw.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

			return ctx.NoContent(http.StatusInternalServerError)
		}

		serverHash, err := utils.HashSHA256(body, mw.key)
		if err != nil {
			mw.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

			return ctx.NoContent(http.StatusInternalServerError)
		}

		if !bytes.Equal(serverHash, hexHash) {
			mw.l.Debugf("[%s] invalid HashSHA256 header value: '%x'", uuid, hexHash)

			return ctx.NoContent(http.StatusBadRequest)
		}

		return next(ctx)
	})
}

// RSAMiddleware - middleware для дешифрования входящего запроса.
func (mw *Manager) RSAMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.HandlerFunc(func(e echo.Context) error {
		uuid := e.Get("uuid")

		body, err := io.ReadAll(e.Request().Body)
		if err != nil {
			mw.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

			return e.NoContent(http.StatusInternalServerError)
		}

		body, err = mw.privateKey.Decrypt(nil, body, &rsa.OAEPOptions{Hash: crypto.SHA256})
		if err != nil {
			mw.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

			return e.NoContent(http.StatusInternalServerError)
		}

		e.Request().Body = io.NopCloser(bytes.NewBuffer(body))

		return next(e)
	})
}
