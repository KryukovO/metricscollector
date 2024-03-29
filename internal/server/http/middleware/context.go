package middleware

import (
	"encoding/hex"
	"encoding/json"

	"github.com/KryukovO/metricscollector/internal/utils"
	"github.com/labstack/echo"
)

// Context реализует интерфейс echo.Context
// и позволяет выполнять подписание ответа хешем.
type Context struct {
	echo.Context
	key []byte
}

// NewContext создает новый контекст.
func NewContext(ctx echo.Context, key []byte) *Context {
	return &Context{
		Context: ctx,
		key:     key,
	}
}

// JSON выполняет отправку ответа в формате JSON с указанным кодом ответа.
func (c *Context) JSON(code int, i interface{}) error {
	if c.key != nil {
		body, err := json.Marshal(i)
		if err != nil {
			return err
		}

		hash, err := utils.HashSHA256(body, c.key)
		if err != nil {
			return err
		}

		c.Response().Header().Add("HashSHA256", hex.EncodeToString(hash))
	}

	return c.Context.JSON(code, i)
}

// HTML выполняет отправку ответа в формате HTML с указанным кодом ответа.
func (c *Context) HTML(code int, html string) error {
	if c.key != nil {
		hash, err := utils.HashSHA256([]byte(html), c.key)
		if err != nil {
			return err
		}

		c.Response().Header().Add("HashSHA256", hex.EncodeToString(hash))
	}

	return c.Context.HTML(code, html)
}
