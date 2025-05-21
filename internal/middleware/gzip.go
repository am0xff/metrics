package middleware

import (
	"bytes"
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
	buf *bytes.Buffer
	key string
}

func newCompressWriter(w http.ResponseWriter, key string) *compressWriter {
	buf := &bytes.Buffer{}
	return &compressWriter{
		w:   w,
		zw:  gzip.NewWriter(buf),
		key: key,
		buf: buf,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
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

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	//return c.zw.Close()
	// 1) Завершаем запись в buf
	if err := c.zw.Close(); err != nil {
		return err
	}
	compressed := c.buf.Bytes()

	if c.key != "" {
		h := utils.CreateHash(compressed, c.key)
		c.w.Header().Set("HashSHA256", h)
	}

	_, err := c.w.Write(compressed)
	return err
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
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GzipMiddleware(next http.Handler, key string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			raw, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()

			cr, err := newCompressReader(io.NopCloser(bytes.NewReader(raw)))
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
