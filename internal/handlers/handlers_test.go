package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/am0xff/metrics/internal/models"
	"github.com/am0xff/metrics/internal/storage"
	memstorage "github.com/am0xff/metrics/internal/storage/memory"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMetric(t *testing.T) {
	ms := memstorage.NewStorage()
	handler := NewHandler(ms)

	ms.Gauges.Set("metric_of_gauge", 1)
	ms.Counters.Set("metric_of_counter", 2)

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
		// Новые тест-кейсы
		{
			name:         "invalid_json_body",
			url:          "/value",
			method:       http.MethodPost,
			body:         `{"invalid": json}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			name:         "empty_id",
			url:          "/value",
			method:       http.MethodPost,
			body:         `{"id": "","type":"gauge"}`,
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
		{
			name:         "empty_type",
			url:          "/value",
			method:       http.MethodPost,
			body:         `{"id": "test","type":""}`,
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
	ms := memstorage.NewStorage()
	handler := NewHandler(ms)

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
		// Новые тест-кейсы
		{
			name:         "invalid_json_body",
			url:          "/update",
			method:       http.MethodPost,
			body:         `{"invalid": json}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			name:         "empty_id",
			url:          "/update",
			method:       http.MethodPost,
			body:         `{"id": "", "type": "gauge", "value": 1.0}`,
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
		{
			name:         "empty_type",
			url:          "/update",
			method:       http.MethodPost,
			body:         `{"id": "test", "type": "", "value": 1.0}`,
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
		{
			name:         "gauge_without_value",
			url:          "/update",
			method:       http.MethodPost,
			body:         `{"id": "test", "type": "gauge"}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			name:         "counter_without_delta",
			url:          "/update",
			method:       http.MethodPost,
			body:         `{"id": "test", "type": "counter"}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			name:         "unknown_metric_type",
			url:          "/update",
			method:       http.MethodPost,
			body:         `{"id": "test", "type": "unknown", "value": 1.0}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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

// Новый тест для POSTUpdatesMetrics
func TestPOSTUpdatesMetrics(t *testing.T) {
	ms := memstorage.NewStorage()
	handler := NewHandler(ms)

	srv := httptest.NewServer(http.HandlerFunc(handler.POSTUpdatesMetrics))
	defer srv.Close()

	testCases := []struct {
		name         string
		method       string
		body         string
		expectedCode int
	}{
		{
			name:         "use_get_method",
			method:       http.MethodGet,
			body:         "",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "valid_batch_update",
			method:       http.MethodPost,
			body:         `[{"id":"cpu","type":"gauge","value":85.5},{"id":"requests","type":"counter","delta":100}]`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid_json",
			method:       http.MethodPost,
			body:         `{"invalid": json}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "empty_array",
			method:       http.MethodPost,
			body:         `[]`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "missing_id",
			method:       http.MethodPost,
			body:         `[{"type":"gauge","value":85.5}]`,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "missing_type",
			method:       http.MethodPost,
			body:         `[{"id":"cpu","value":85.5}]`,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "gauge_without_value",
			method:       http.MethodPost,
			body:         `[{"id":"cpu","type":"gauge"}]`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "counter_without_delta",
			method:       http.MethodPost,
			body:         `[{"id":"requests","type":"counter"}]`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "unknown_metric_type",
			method:       http.MethodPost,
			body:         `[{"id":"test","type":"unknown","value":1.0}]`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tc.method
			req.URL = srv.URL

			if len(tc.body) > 0 {
				req.SetHeader("Content-Type", "application/json")
				req.SetBody(tc.body)
			}

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Response code didn't match expected")
		})
	}
}

// Тест для GETGetMetric с роутером
func TestGETGetMetric(t *testing.T) {
	ms := memstorage.NewStorage()
	ms.SetGauge(context.Background(), "cpu", storage.Gauge(85.5))
	ms.SetCounter(context.Background(), "requests", storage.Counter(100))

	r := chi.NewRouter()
	handler := NewHandler(ms)
	r.Get("/value/{type}/{name}", handler.GETGetMetric)

	srv := httptest.NewServer(r)
	defer srv.Close()

	testCases := []struct {
		name         string
		url          string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "get_gauge_metric",
			url:          "/value/gauge/cpu",
			expectedCode: http.StatusOK,
			expectedBody: "85.5",
		},
		{
			name:         "get_counter_metric",
			url:          "/value/counter/requests",
			expectedCode: http.StatusOK,
			expectedBody: "100",
		},
		{
			name:         "get_nonexistent_gauge",
			url:          "/value/gauge/nonexistent",
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
		{
			name:         "get_nonexistent_counter",
			url:          "/value/counter/nonexistent",
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
		{
			name:         "invalid_metric_type",
			url:          "/value/unknown/cpu",
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(srv.URL + tc.url)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedCode, resp.StatusCode)

			if tc.expectedBody != "" {
				body := make([]byte, len(tc.expectedBody))
				resp.Body.Read(body)
				assert.Equal(t, tc.expectedBody, string(body))
			}
		})
	}
}

// Тест для GETUpdateMetric
func TestGETUpdateMetric(t *testing.T) {
	ms := memstorage.NewStorage()

	r := chi.NewRouter()
	handler := NewHandler(ms)
	r.Post("/update/{type}/{name}/{value}", handler.GETUpdateMetric)

	srv := httptest.NewServer(r)
	defer srv.Close()

	testCases := []struct {
		name         string
		url          string
		expectedCode int
	}{
		{
			name:         "update_gauge_metric",
			url:          "/update/gauge/cpu/85.5",
			expectedCode: http.StatusOK,
		},
		{
			name:         "update_counter_metric",
			url:          "/update/counter/requests/100",
			expectedCode: http.StatusOK,
		},
		{
			name:         "empty_name",
			url:          "/update/gauge//85.5",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "invalid_gauge_value",
			url:          "/update/gauge/cpu/invalid",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid_counter_value",
			url:          "/update/counter/requests/invalid",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "unknown_metric_type",
			url:          "/update/unknown/test/123",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Post(srv.URL+tc.url, "", nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedCode, resp.StatusCode)
		})
	}
}

// Тест для GetMetrics (HTML страница)
func TestGetMetrics(t *testing.T) {
	ms := memstorage.NewStorage()
	ms.SetGauge(context.Background(), "cpu", storage.Gauge(85.5))
	ms.SetCounter(context.Background(), "requests", storage.Counter(100))

	handler := NewHandler(ms)
	srv := httptest.NewServer(http.HandlerFunc(handler.GetMetrics))
	defer srv.Close()

	testCases := []struct {
		name         string
		method       string
		expectedCode int
		checkContent bool
	}{
		{
			name:         "get_metrics_page",
			method:       http.MethodGet,
			expectedCode: http.StatusOK,
			checkContent: true,
		},
		{
			name:         "post_method_not_allowed",
			method:       http.MethodPost,
			expectedCode: http.StatusMethodNotAllowed,
			checkContent: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, srv.URL, nil)
			require.NoError(t, err)

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedCode, resp.StatusCode)

			if tc.checkContent {
				assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))

				body := new(bytes.Buffer)
				body.ReadFrom(resp.Body)
				content := body.String()

				// Проверяем, что HTML содержит наши метрики
				assert.Contains(t, content, "<html>")
				assert.Contains(t, content, "cpu")
				assert.Contains(t, content, "requests")
				assert.Contains(t, content, "85.5")
				assert.Contains(t, content, "100")
			}
		})
	}
}

// Тест для Ping
func TestPing(t *testing.T) {
	ms := memstorage.NewStorage()
	handler := NewHandler(ms)

	srv := httptest.NewServer(http.HandlerFunc(handler.Ping))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Тест для NewHandler
func TestNewHandler(t *testing.T) {
	ms := memstorage.NewStorage()
	handler := NewHandler(ms)

	assert.NotNil(t, handler)
	assert.Equal(t, ms, handler.storageProvider)
}

// Интеграционный тест с полным роутером
func TestFullRouterIntegration(t *testing.T) {
	ms := memstorage.NewStorage()

	// Создаем роутер вручную (избегаем циклического импорта)
	r := chi.NewRouter()
	handler := NewHandler(ms)
	r.Get("/", handler.GetMetrics)
	r.Get("/ping", handler.Ping)
	r.Post("/value/", handler.POSTGetMetric)
	r.Post("/update/", handler.POSTUpdateMetric)
	r.Post("/updates/", handler.POSTUpdatesMetrics)
	r.Get("/value/{type}/{name}", handler.GETGetMetric)
	r.Post("/update/{type}/{name}/{value}", handler.GETUpdateMetric)

	srv := httptest.NewServer(r)
	defer srv.Close()

	client := resty.New()

	// Тест обновления gauge метрики через JSON
	gaugeMetric := models.Metrics{
		ID:    "cpu_usage",
		MType: storage.MetricTypeGauge,
		Value: func() *float64 { v := 85.5; return &v }(),
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(gaugeMetric).
		Post(srv.URL + "/update/")

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	// Тест получения метрики через JSON
	getMetric := models.Metrics{
		ID:    "cpu_usage",
		MType: storage.MetricTypeGauge,
	}

	resp, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(getMetric).
		Post(srv.URL + "/value/")

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	var response models.Metrics
	err = json.Unmarshal(resp.Body(), &response)
	require.NoError(t, err)
	assert.Equal(t, "cpu_usage", response.ID)
	assert.Equal(t, storage.MetricTypeGauge, response.MType)
	assert.NotNil(t, response.Value)
	assert.Equal(t, 85.5, *response.Value)

	// Тест массового обновления
	batchMetrics := []models.Metrics{
		{
			ID:    "memory_usage",
			MType: storage.MetricTypeGauge,
			Value: func() *float64 { v := 67.2; return &v }(),
		},
		{
			ID:    "requests_total",
			MType: storage.MetricTypeCounter,
			Delta: func() *int64 { v := int64(1000); return &v }(),
		},
	}

	resp, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(batchMetrics).
		Post(srv.URL + "/updates/")

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	// Тест получения метрики через URL
	resp, err = client.R().Get(srv.URL + "/value/gauge/memory_usage")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "67.2", string(resp.Body()))

	// Тест обновления через URL
	resp, err = client.R().Post(srv.URL + "/update/counter/errors/5")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	// Проверяем, что метрика сохранилась
	resp, err = client.R().Get(srv.URL + "/value/counter/errors")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "5", string(resp.Body()))

	// Тест HTML страницы
	resp, err = client.R().Get(srv.URL + "/")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Contains(t, string(resp.Body()), "<html>")
	assert.Contains(t, string(resp.Body()), "cpu_usage")
	assert.Contains(t, string(resp.Body()), "memory_usage")
	assert.Contains(t, string(resp.Body()), "requests_total")

	// Тест ping
	resp, err = client.R().Get(srv.URL + "/ping")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

// Тест для проверки типов метрик в models
func TestModelsMetricsString(t *testing.T) {
	testCases := []struct {
		name     string
		metric   models.Metrics
		expected string
	}{
		{
			name: "gauge_metric",
			metric: models.Metrics{
				ID:    "cpu",
				MType: storage.MetricTypeGauge,
				Value: func() *float64 { v := 85.5; return &v }(),
			},
			expected: "85.5",
		},
		{
			name: "counter_metric",
			metric: models.Metrics{
				ID:    "requests",
				MType: storage.MetricTypeCounter,
				Delta: func() *int64 { v := int64(100); return &v }(),
			},
			expected: "100",
		},
		{
			name: "gauge_nil_value",
			metric: models.Metrics{
				ID:    "cpu",
				MType: storage.MetricTypeGauge,
				Value: nil,
			},
			expected: "",
		},
		{
			name: "counter_nil_delta",
			metric: models.Metrics{
				ID:    "requests",
				MType: storage.MetricTypeCounter,
				Delta: nil,
			},
			expected: "",
		},
		{
			name: "unknown_type",
			metric: models.Metrics{
				ID:    "unknown",
				MType: "unknown",
			},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.metric.String()
			assert.Equal(t, tc.expected, result)
		})
	}
}
