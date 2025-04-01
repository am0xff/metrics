package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/am0xff/metrics/internal/handlers"
	"github.com/am0xff/metrics/internal/storage"
)

func TestHandler_RootHandle(t *testing.T) {
	tests := []struct {
		name            string
		storageSetup    func() *storage.MemStorage
		expectedSubstrs []string
	}{
		{
			name: "Empty storage",
			storageSetup: func() *storage.MemStorage {
				return storage.NewMemStorage()
			},
			expectedSubstrs: []string{
				"<html>",
				"<head><title>Metrics</title></head>",
				"<ul></ul>",
			},
		},
		{
			name: "Non-empty storage",
			storageSetup: func() *storage.MemStorage {
				s := storage.NewMemStorage()
				s.SetGauge("g1", 1.23)
				s.SetCounter("c1", 10)
				return s
			},
			expectedSubstrs: []string{
				"g1: 1.23",
				"c1: 10",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := tt.storageSetup()
			h := handlers.NewHandler(store)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			h.RootHandle(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Status should be %d, but get %d", http.StatusOK, rec.Code)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != "text/html; charset=utf-8" {
				t.Errorf("Content-Type should be 'text/html; charset=utf-8', get %q", contentType)
			}

			body := rec.Body.String()
			for _, substr := range tt.expectedSubstrs {
				if !strings.Contains(body, substr) {
					t.Errorf("Body should be, %q, but get: %s", substr, body)
				}
			}
		})
	}
}
