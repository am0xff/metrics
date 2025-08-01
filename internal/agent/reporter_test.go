package agent

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/am0xff/metrics/internal/models"
	"github.com/am0xff/metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReporter(t *testing.T) {
	cfg := &ReporterConfig{
		ServerAddr: "localhost:8080",
		Key:        "secret",
	}

	reporter := NewReporter(cfg)

	assert.NotNil(t, reporter)
	assert.Equal(t, cfg, reporter.cfg)
	assert.NotNil(t, reporter.client)
}

func TestReporter_Report(t *testing.T) {
	// Создаем тестовый HTTP сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем заголовки
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
		assert.Equal(t, http.MethodPost, r.Method)

		// Проверяем путь
		expectedPaths := []string{"/update/"}
		found := false
		for _, path := range expectedPaths {
			if r.URL.Path == path {
				found = true
				break
			}
		}
		assert.True(t, found, "Unexpected path: %s", r.URL.Path)

		// Читаем и разжимаем тело запроса
		gz, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer gz.Close()

		body, err := io.ReadAll(gz)
		require.NoError(t, err)

		// Парсим JSON
		var metric models.Metrics
		err = json.Unmarshal(body, &metric)
		require.NoError(t, err)

		// Проверяем содержимое метрики
		assert.NotEmpty(t, metric.ID)
		assert.NotEmpty(t, metric.MType)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Настраиваем reporter
	cfg := &ReporterConfig{
		ServerAddr: server.URL[7:], // убираем "http://"
		Key:        "",
	}
	reporter := NewReporter(cfg)

	// Тестовые данные
	gauges := map[string]float64{
		"cpu_usage":    85.5,
		"memory_usage": 67.2,
	}
	counters := map[string]int64{
		"requests_total": 1000,
		"errors_total":   5,
	}

	// Вызываем Report - не должно быть паник или ошибок
	reporter.Report(gauges, counters)
}

func TestReporter_ReportBatch(t *testing.T) {
	// Создаем тестовый HTTP сервер для batch запросов
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем заголовки
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/updates/", r.URL.Path)

		// Читаем и разжимаем тело запроса
		gz, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer gz.Close()

		body, err := io.ReadAll(gz)
		require.NoError(t, err)

		// Парсим JSON массив
		var metrics []models.Metrics
		err = json.Unmarshal(body, &metrics)
		require.NoError(t, err)

		// Проверяем, что получили массив метрик
		assert.NotEmpty(t, metrics)

		// Проверяем типы метрик
		for _, metric := range metrics {
			assert.NotEmpty(t, metric.ID)
			assert.Contains(t, []storage.MetricType{storage.MetricTypeGauge, storage.MetricTypeCounter}, metric.MType)

			if metric.MType == storage.MetricTypeGauge {
				assert.NotNil(t, metric.Value)
			} else if metric.MType == storage.MetricTypeCounter {
				assert.NotNil(t, metric.Delta)
			}
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Настраиваем reporter
	cfg := &ReporterConfig{
		ServerAddr: server.URL[7:], // убираем "http://"
		Key:        "",
	}
	reporter := NewReporter(cfg)

	// Тестовые данные
	gauges := map[string]float64{
		"cpu_usage":    85.5,
		"memory_usage": 67.2,
	}
	counters := map[string]int64{
		"requests_total": 1000,
		"errors_total":   5,
	}

	// Вызываем ReportBatch
	reporter.ReportBatch(gauges, counters)
}

func TestReporter_ReportWithKey(t *testing.T) {
	// Создаем тестовый HTTP сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем наличие хеша в заголовке
		hash := r.Header.Get("HashSHA256")
		assert.NotEmpty(t, hash, "Expected hash in header when key is provided")

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Настраиваем reporter с ключом
	cfg := &ReporterConfig{
		ServerAddr: server.URL[7:], // убираем "http://"
		Key:        "secret_key",
	}
	reporter := NewReporter(cfg)

	// Тестовые данные
	gauges := map[string]float64{"cpu": 85.5}
	counters := map[string]int64{}

	// Вызываем Report с ключом
	reporter.Report(gauges, counters)
}

func TestReporter_ReportEmpty(t *testing.T) {
	// Создаем тестовый HTTP сервер
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Настраиваем reporter
	cfg := &ReporterConfig{
		ServerAddr: server.URL[7:], // убираем "http://"
		Key:        "",
	}
	reporter := NewReporter(cfg)

	// Пустые данные
	gauges := map[string]float64{}
	counters := map[string]int64{}

	// Вызываем Report с пустыми данными
	reporter.Report(gauges, counters)

	// Не должно быть запросов при пустых данных
	assert.Equal(t, 0, requestCount)
}

func TestReporter_ReportBatchEmpty(t *testing.T) {
	// Создаем тестовый HTTP сервер
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// Если запрос все же пришел, проверим что это пустой массив
		gz, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer gz.Close()

		body, err := io.ReadAll(gz)
		require.NoError(t, err)

		var metrics []models.Metrics
		err = json.Unmarshal(body, &metrics)
		require.NoError(t, err)
		assert.Empty(t, metrics)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Настраиваем reporter
	cfg := &ReporterConfig{
		ServerAddr: server.URL[7:], // убираем "http://"
		Key:        "",
	}
	reporter := NewReporter(cfg)

	// Пустые данные
	gauges := map[string]float64{}
	counters := map[string]int64{}

	// Вызываем ReportBatch с пустыми данными
	reporter.ReportBatch(gauges, counters)

	// Запрос все равно должен отправиться (даже с пустым массивом)
	assert.Equal(t, 1, requestCount)
}

func TestReporter_ServerError(t *testing.T) {
	// Создаем тестовый HTTP сервер, возвращающий ошибку
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Настраиваем reporter
	cfg := &ReporterConfig{
		ServerAddr: server.URL[7:], // убираем "http://"
		Key:        "",
	}
	reporter := NewReporter(cfg)

	// Тестовые данные
	gauges := map[string]float64{"cpu": 85.5}
	counters := map[string]int64{}

	// Вызываем Report - не должно быть паник, только логи
	reporter.Report(gauges, counters)
}

func TestReporterConfig(t *testing.T) {
	// Тестируем структуру конфигурации
	cfg := &ReporterConfig{
		ServerAddr: "localhost:8080",
		Key:        "secret_key",
	}

	assert.Equal(t, "localhost:8080", cfg.ServerAddr)
	assert.Equal(t, "secret_key", cfg.Key)
}
