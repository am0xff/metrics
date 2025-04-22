package handlers

import (
	"github.com/am0xff/metrics/internal/storage"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMetric(t *testing.T) {
	s := storage.NewMemStorage()
	handler := NewHandler(s)

	s.Gauges.Set("metric_of_gauge", 1)
	s.Counters.Set("metric_of_counter", 2)

	h := http.HandlerFunc(handler.POSTGetMetric)
	srv := httptest.NewServer(h)
	defer srv.Close()

	testCases := []struct {
		name         string
		url          string
		method       string
		body         string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "use_get_method",
			url:          "/value",
			method:       http.MethodGet,
			body:         "",
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{
			name:         "get_with_name_and_type_gauge",
			url:          "/value",
			method:       http.MethodPost,
			body:         `{"id": "metric_of_gauge","type":"gauge"}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id": "metric_of_gauge","type":"gauge","value":1}`,
		},
		{
			name:         "get_with_name_and_type_counter",
			url:          "/value",
			method:       http.MethodPost,
			body:         `{"id": "metric_of_counter","type":"counter"}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id": "metric_of_counter","type":"counter","delta":2}`,
		},
		{
			name:         "get_with_name_and_type_unknown",
			url:          "/value",
			method:       http.MethodPost,
			body:         `{"id": "metric_of_counter","type":"unknown"}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			name:         "get_with_incorrect_name_and_type_gauge",
			url:          "/value",
			method:       http.MethodPost,
			body:         `{"id": "unknown","type":"gauge"}`,
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
		{
			name:         "get_with_incorrect_name_and_type_counter",
			url:          "/value",
			method:       http.MethodPost,
			body:         `{"id": "unknown","type":"counter"}`,
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tc.method
			req.URL = srv.URL + tc.url

			if len(tc.body) > 0 {
				req.SetHeader("Content-Type", "application/json")
				req.SetBody(tc.body)
			}

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Response code didn't match expected")
			// проверяем корректность полученного тела ответа, если мы его ожидаем
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, string(resp.Body()))
			}
		})
	}
}

func TestUpdateMetric(t *testing.T) {
	s := storage.NewMemStorage()
	handler := NewHandler(s)

	h := http.HandlerFunc(handler.POSTUpdateMetric)
	srv := httptest.NewServer(h)
	defer srv.Close()

	testCases := []struct {
		name         string
		url          string
		method       string
		body         string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "use_get_method",
			url:          "/update",
			method:       http.MethodGet,
			body:         "",
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{
			name:         "update_with_name_and_type_gauge",
			url:          "/update",
			method:       http.MethodPost,
			body:         `{"id": "metric_of_gauge", "type": "gauge", "value": 1.0}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id": "metric_of_gauge", "type": "gauge", "value": 1.0}`,
		},
		{
			name:         "update_with_name_and_type_counter",
			url:          "/update",
			method:       http.MethodPost,
			body:         `{"id": "metric_of_counter", "type": "counter", "delta": 1}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id": "metric_of_counter", "type": "counter", "delta": 1}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tc.method
			req.URL = srv.URL + tc.url

			if len(tc.body) > 0 {
				req.SetHeader("Content-Type", "application/json")
				req.SetBody(tc.body)
			}

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Response code didn't match expected")
			// проверяем корректность полученного тела ответа, если мы его ожидаем
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, string(resp.Body()))
			}
		})
	}
}
