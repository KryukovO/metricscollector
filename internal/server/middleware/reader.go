package middleware

import (
	"compress/gzip"
	"io"
)

// CompressReader реализует интерфейс io.ReadCloser
// и позволяет декомпрессировать получаемые данные, используя gzip.
type CompressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewCompressReader создаёт новый объект CompressReader.
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

// Read выполняет чтение несжатых байт.
func (c CompressReader) Read(p []byte) (int, error) {
	return c.zr.Read(p)
}

// Close pакрывает CompressReader.
func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}

	return c.zr.Close()
}
