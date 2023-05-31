package models

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// CompressWriter реализует интерфейс http.ResponseWriter
// и выполняет сжатие данных, используя gzip, если это необходимо
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

// CompressReader реализует интерфейс io.ReadCloser
// и позволяет декомпрессировать получаемые данные, используя gzip
type CompressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &CompressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c CompressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
