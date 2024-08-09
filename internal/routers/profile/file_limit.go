package profile

import (
	"fmt"
	"io"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/bytes"
)

func isOverSizeLimit(bodyLimit string) (bool, error) {
	limit, err := bytes.Parse(bodyLimit)
	if err != nil {
		panic(fmt.Errorf("echo: invalid body-limit=%s", limit))
	}
	// config.limit = limit
	pool := limitedReaderPool(config)

	// Based on content read
	r := pool.Get().(*limitedReader)
	r.Reset(req.Body)
	defer pool.Put(r)
	req.Body = r
}

// BodyLimitConfig defines the config for BodyLimit middleware.
type BodyLimitConfig struct {
	// Maximum allowed size for a request body, it can be specified
	// as `4x` or `4xB`, where x is one of the multiple from K, M, G, T or P.
	Limit string `yaml:"limit"`
	limit int64
}

type limitedReader struct {
	BodyLimitConfig
	reader io.ReadCloser
	read   int64
}

func (r *limitedReader) Read(b []byte) (n int, err error) {
	n, err = r.reader.Read(b)
	r.read += int64(n)
	if r.read > r.limit {
		return n, echo.ErrStatusRequestEntityTooLarge
	}
	return
}

func (r *limitedReader) Close() error {
	return r.reader.Close()
}

func (r *limitedReader) Reset(reader io.ReadCloser) {
	r.reader = reader
	r.read = 0
}

func limitedReaderPool(c BodyLimitConfig) sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			return &limitedReader{BodyLimitConfig: c}
		},
	}
}
