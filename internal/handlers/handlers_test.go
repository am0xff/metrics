package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/am0xff/metrics/internal/handlers"
	"github.com/am0xff/metrics/internal/storage"
)

func TestAPIUpdate(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		contentType        string
		url                string
		expectedStatus     int
		expectGaugeValue   float64
		expectCounterValue int64
	}{
		{
			name:           "Метод не POST",
			method:         http.MethodGet,
			contentType:    "text/plain",
			url:            "/update/gauge/test/100",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Неверный Content-Type",
			method:         http.MethodPost,
			contentType:    "application/json",
			url:            "/update/gauge/test/100",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Неверное число частей URL",
			method:         http.MethodPost,
			contentType:    "text/plain",
			url:            "/update/gauge/test", // ожидается 5 частей, а здесь их 4
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Пустое имя метрики",
			method:         http.MethodPost,
			contentType:    "text/plain",
			url:            "/update/gauge//100", // имя отсутствует
			expectedStatus: http.StatusNotFound,
		},
		{
			name:               "Gauge корректное значение",
			method:             http.MethodPost,
			contentType:        "text/plain",
			url:                "/update/gauge/testGauge/123.456",
			expectedStatus:     http.StatusOK,
			expectGaugeValue:   float64(123.456),
			expectCounterValue: 0,
		},
		{
			name:           "Gauge некорректное значение",
			method:         http.MethodPost,
			contentType:    "text/plain",
			url:            "/update/gauge/testGauge/notafloat",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:               "Counter корректное значение",
			method:             http.MethodPost,
			contentType:        "text/plain",
			url:                "/update/counter/testCounter/789",
			expectedStatus:     http.StatusOK,
			expectGaugeValue:   0,
			expectCounterValue: int64(789),
		},
		{
			name:           "Counter некорректное значение",
			method:         http.MethodPost,
			contentType:    "text/plain",
			url:            "/update/counter/testCounter/notanint",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Неизвестный тип метрики",
			method:         http.MethodPost,
			contentType:    "text/plain",
			url:            "/update/unknown/testCounter/100",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			handler := handlers.NewHandler(store)

			req := httptest.NewRequest(tc.method, tc.url, nil)
			req.Header.Set("Content-Type", tc.contentType)
			rec := httptest.NewRecorder()

			handler.APIUpdate(rec, req)
			if rec.Code != tc.expectedStatus {
				t.Errorf("Ожидался статус %d, получен %d", tc.expectedStatus, rec.Code)
			}

			// Если запрос корректный, проверяем обновление хранилища.
			if rec.Code == http.StatusOK {
				parts := strings.Split(req.URL.Path, "/")
				metricType := parts[2]
				metricName := parts[3]
				switch metricType {
				case "gauge":
					val, ok := store.GetGauge(metricName)
					if !ok {
						t.Errorf("Gauge %s is not found", metricName)
					} else if float64(val) != tc.expectGaugeValue {
						t.Errorf("Gauge should be %s = %v, but got %v", metricName, tc.expectGaugeValue, val)
					}
				case "counter":
					val, ok := store.GetCounter(metricName)
					if !ok {
						t.Errorf("Counter %s is not found", metricName)
					} else if int64(val) != tc.expectCounterValue {
						t.Errorf("Counter should be %s = %v, but got %v", metricName, tc.expectGaugeValue, val)
					}
				}
			}
		})
	}
}
