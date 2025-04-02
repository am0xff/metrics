package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/am0xff/metrics/internal/storage"
)

func TestListMetricsHandler(t *testing.T) {
	testCases := []struct {
		name               string
		gauges             map[string]float64
		counters           map[string]int64
		expectedSubstrings []string
	}{
		{
			name:               "empty storage",
			gauges:             nil,
			counters:           nil,
			expectedSubstrings: []string{"<ul></ul>"},
		},
		{
			name:               "only gauge",
			gauges:             map[string]float64{"Alloc": 100},
			counters:           nil,
			expectedSubstrings: []string{"<li>Alloc: 100</li>"},
		},
		{
			name:               "only counter",
			gauges:             nil,
			counters:           map[string]int64{"PollCount": 10},
			expectedSubstrings: []string{"<li>PollCount: 10</li>"},
		},
		{
			name:               "both metrics",
			gauges:             map[string]float64{"Alloc": 100},
			counters:           map[string]int64{"PollCount": 10},
			expectedSubstrings: []string{"<li>Alloc: 100</li>", "<li>PollCount: 10</li>"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			if tc.gauges != nil {
				for k, v := range tc.gauges {
					store.UpdateGauge(k, v)
				}
			}
			if tc.counters != nil {
				for k, v := range tc.counters {
					store.UpdateCounter(k, v)
				}
			}

			srv := NewServer(store)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			srv.ListMetricsHandler(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
			}

			if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
				t.Errorf("expected Content-Type %q, got %q", "text/html; charset=utf-8", ct)
			}

			body := rr.Body.String()
			for _, substr := range tc.expectedSubstrings {
				if !strings.Contains(body, substr) {
					t.Errorf("response body does not contain expected substring %q: %s", substr, body)
				}
			}
		})
	}
}
