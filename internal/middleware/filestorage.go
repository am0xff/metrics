package middleware

import (
	"fmt"
	"github.com/am0xff/metrics/internal/storage"
	"net/http"
	"strings"
)

func FileStorageMiddleware(ms *storage.MemStorage, filename string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)

		if r.Method == http.MethodPost && strings.Contains(r.RequestURI, "update") {
			if err := ms.Save(filename); err != nil {
				fmt.Println("save storage: %w", err)
			}
		}
	})
}
