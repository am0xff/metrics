package middleware

import (
	"compress/gzip"
	"github.com/am0xff/metrics/internal/utils"
	"io"
	"net/http"
	"strings"
)

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	w   http.ResponseWriter
	zw  *gzip.Writer
	key string
}

func newCompressWriter(w http.ResponseWriter, key string) *compressWriter {
	return &compressWriter{
		w:   w,
		zw:  gzip.NewWriter(w),
		key: key,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if c.key != "" {
		h := utils.CreateHash(p, c.key)
		c.w.Header().Set("HashSHA256", h)
	}

	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	contentType := c.w.Header().Get("Content-Type")
	isJSONOrHTML := strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")
	if statusCode < 300 && isJSONOrHTML {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer
func (c *compressWriter) Close() error {
	if c.zw == nil {
		return nil
	}
	return c.zw.Close()
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	var err error
	if c.r != nil {
		err = c.r.Close()
	}

	if c.zr != nil {
		if closeErr := c.zr.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}

	return err
}

func GzipMiddleware(next http.Handler, key string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw := newCompressWriter(w, key)
			ow = cw
			defer cw.Close()
		}

		next.ServeHTTP(ow, r)
	})
}
