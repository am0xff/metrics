package middleware

import (
	"bytes"
	"github.com/am0xff/metrics/internal/utils"
	"io"
	"net/http"
)

func HashMiddleware(next http.Handler, key string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if key != "" {
			sig := r.Header.Get("HashSHA256")
			if sig != "" {
				raw, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				_ = r.Body.Close()

				r.Body = io.NopCloser(bytes.NewReader(raw))

				if err := utils.ValidateHash(raw, key, sig); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}
