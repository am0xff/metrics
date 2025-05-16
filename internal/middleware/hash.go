package middleware

import (
	"bytes"
	"github.com/am0xff/metrics/internal/utils"
	"io"
	"net/http"
	"strings"
)

func HashMiddleware(h http.Handler, key string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hashHeader := r.Header.Get("HashSHA256")

		if hashHeader == "" || key == "" {
			h.ServeHTTP(w, r)
			return
		}

		sendsGzip := strings.Contains(r.Header.Get("Content-Encoding"), "gzip")
		if !sendsGzip {
			h.ServeHTTP(w, r)
			return
		}

		var bodyBytes []byte
		if r.Body != nil {
			data, err := io.ReadAll(r.Body)
			if err == nil {
				bodyBytes = data
			}
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		if len(bodyBytes) == 0 {
			h.ServeHTTP(w, r)
			return
		}

		if err := utils.ValidateHash(bodyBytes, key, hashHeader); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		h.ServeHTTP(w, r)
	})
}
