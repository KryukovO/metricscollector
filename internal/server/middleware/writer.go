package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// CompressWriter реализует интерфейс http.ResponseWriter
// и выполняет сжатие данных, используя gzip, если это необходимо.
type CompressWriter struct {
	http.ResponseWriter
	zw          *gzip.Writer
	acceptTypes []string // Допустимые для сжатия форматы тела ответа
}

func NewCompressWriter(w http.ResponseWriter, acceptTypes []string) *CompressWriter {
	return &CompressWriter{
		ResponseWriter: w,
		zw:             gzip.NewWriter(w),
		acceptTypes:    acceptTypes,
	}
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	defer c.Close()

	// Проверка нужно ли сжимать данные
	contentType := c.Header().Get("Content-Type")
	for _, t := range c.acceptTypes {
		if strings.Contains(contentType, t) {
			return c.zw.Write(p)
		}
	}

	return c.ResponseWriter.Write(p)
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	// Добавляем заголовок с информацией о сжатии только,
	// если формат тела ответа допустим для сжатия
	contentType := c.Header().Get("Content-Type")
	for _, t := range c.acceptTypes {
		if strings.Contains(contentType, t) {
			c.Header().Set("Content-Encoding", "gzip")

			break
		}
	}

	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	return c.zw.Close()
}
