package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/am0xff/metrics/internal/handlers"
	"github.com/am0xff/metrics/internal/models"
	"github.com/am0xff/metrics/internal/storage"
	memstorage "github.com/am0xff/metrics/internal/storage/memory"
	"github.com/go-chi/chi/v5"
)

// ExampleNewHandler демонстрирует создание нового обработчика метрик
func ExampleNewHandler() {
	// Создаем хранилище в памяти
	s := memstorage.NewStorage()

	// Создаем обработчик
	handler := handlers.NewHandler(s)

	fmt.Printf("Handler created: %T\n", handler)
	// Output: Handler created: *handlers.Handler
}

// ExampleHandler_POSTUpdateMetric демонстрирует обновление gauge метрики через JSON API
func ExampleHandler_POSTUpdateMetric_gauge() {
	// Настройка
	s := memstorage.NewStorage()
	handler := handlers.NewHandler(s)

	// Создаем HTTP сервер для тестирования
	server := httptest.NewServer(http.HandlerFunc(handler.POSTUpdateMetric))
	defer server.Close()

	// Подготавливаем данные для gauge метрики
	metric := models.Metrics{
		ID:    "cpu_usage",
		MType: storage.MetricTypeGauge,
		Value: func() *float64 { v := 85.5; return &v }(),
	}

	jsonData, _ := json.Marshal(metric)

	// Отправляем POST запрос
	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, _ := io.ReadAll(resp.Body)
	var response models.Metrics
	json.Unmarshal(body, &response)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Metric ID: %s\n", response.ID)
	fmt.Printf("Metric Type: %s\n", response.MType)
	fmt.Printf("Metric Value: %.1f\n", *response.Value)

	// Output:
	// Status: 200
	// Metric ID: cpu_usage
	// Metric Type: gauge
	// Metric Value: 85.5
}

// ExampleHandler_POSTUpdateMetric_counter демонстрирует обновление counter метрики
func ExampleHandler_POSTUpdateMetric_counter() {
	s := memstorage.NewStorage()
	handler := handlers.NewHandler(s)

	server := httptest.NewServer(http.HandlerFunc(handler.POSTUpdateMetric))
	defer server.Close()

	// Подготавливаем данные для counter метрики
	metric := models.Metrics{
		ID:    "requests_total",
		MType: storage.MetricTypeCounter,
		Delta: func() *int64 { v := int64(100); return &v }(),
	}

	jsonData, _ := json.Marshal(metric)

	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var response models.Metrics
	json.Unmarshal(body, &response)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Metric ID: %s\n", response.ID)
	fmt.Printf("Metric Type: %s\n", response.MType)
	fmt.Printf("Metric Delta: %d\n", *response.Delta)

	// Output:
	// Status: 200
	// Metric ID: requests_total
	// Metric Type: counter
	// Metric Delta: 100
}

// ExampleHandler_POSTGetMetric демонстрирует получение метрики через JSON API
func ExampleHandler_POSTGetMetric() {
	s := memstorage.NewStorage()
	handler := handlers.NewHandler(s)

	// Предварительно добавляем метрику
	s.SetGauge(context.Background(), "memory_usage", storage.Gauge(67.2))

	server := httptest.NewServer(http.HandlerFunc(handler.POSTGetMetric))
	defer server.Close()

	// Запрос метрики
	requestMetric := models.Metrics{
		ID:    "memory_usage",
		MType: storage.MetricTypeGauge,
	}

	jsonData, _ := json.Marshal(requestMetric)
	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var response models.Metrics
	json.Unmarshal(body, &response)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Retrieved metric: %s = %.1f\n", response.ID, *response.Value)

	// Output:
	// Status: 200
	// Retrieved metric: memory_usage = 67.2
}

// ExampleHandler_POSTUpdatesMetrics демонстрирует массовое обновление метрик
func ExampleHandler_POSTUpdatesMetrics() {
	s := memstorage.NewStorage()
	handler := handlers.NewHandler(s)

	server := httptest.NewServer(http.HandlerFunc(handler.POSTUpdatesMetrics))
	defer server.Close()

	// Подготавливаем массив метрик
	metrics := []models.Metrics{
		{
			ID:    "cpu_usage",
			MType: storage.MetricTypeGauge,
			Value: func() *float64 { v := 85.5; return &v }(),
		},
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

	jsonData, _ := json.Marshal(metrics)
	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Batch update status: %d\n", resp.StatusCode)

	// Проверяем, что метрики сохранились
	if cpu, ok := s.GetGauge(context.Background(), "cpu_usage"); ok {
		fmt.Printf("CPU Usage: %.1f\n", float64(cpu))
	}
	if requests, ok := s.GetCounter(context.Background(), "requests_total"); ok {
		fmt.Printf("Requests Total: %d\n", int64(requests))
	}

	// Output:
	// Batch update status: 200
	// CPU Usage: 85.5
	// Requests Total: 1000
}

// ExampleHandler_GETGetMetric демонстрирует получение метрики через URL параметры
func ExampleHandler_GETGetMetric() {
	s := memstorage.NewStorage()

	// Добавляем тестовые метрики
	s.SetGauge(context.Background(), "temperature", storage.Gauge(23.7))
	s.SetCounter(context.Background(), "errors", storage.Counter(42))

	// Создаем роутер вручную (без импорта router пакета)
	r := chi.NewRouter()
	handler := handlers.NewHandler(s)
	r.Get("/value/{type}/{name}", handler.GETGetMetric)

	server := httptest.NewServer(r)
	defer server.Close()

	// Получаем gauge метрику
	resp1, err := http.Get(server.URL + "/value/gauge/temperature")
	if err != nil {
		log.Fatal(err)
	}
	defer resp1.Body.Close()

	body1, _ := io.ReadAll(resp1.Body)
	fmt.Printf("Temperature: %s\n", string(body1))

	// Получаем counter метрику
	resp2, err := http.Get(server.URL + "/value/counter/errors")
	if err != nil {
		log.Fatal(err)
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)
	fmt.Printf("Errors: %s\n", string(body2))

	// Output:
	// Temperature: 23.7
	// Errors: 42
}

// ExampleHandler_GetMetrics демонстрирует получение всех метрик в HTML формате
func ExampleHandler_GetMetrics() {
	s := memstorage.NewStorage()
	handler := handlers.NewHandler(s)

	// Добавляем несколько метрик
	s.SetGauge(context.Background(), "cpu", storage.Gauge(85.5))
	s.SetCounter(context.Background(), "requests", storage.Counter(1000))

	server := httptest.NewServer(http.HandlerFunc(handler.GetMetrics))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("Status: %d\n", resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)

	// Проверяем, что HTML содержит наши метрики
	if bytes.Contains(body, []byte("cpu")) && bytes.Contains(body, []byte("requests")) {
		fmt.Println("HTML contains metrics: ✓")
	}

	// Output:
	// Content-Type: text/html
	// Status: 200
	// HTML contains metrics: ✓
}

// ExampleHandler_Ping демонстрирует проверку доступности хранилища
func ExampleHandler_Ping() {
	s := memstorage.NewStorage()
	handler := handlers.NewHandler(s)

	server := httptest.NewServer(http.HandlerFunc(handler.Ping))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Ping status: %d\n", resp.StatusCode)

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Storage is available: ✓")
	}

	// Output:
	// Ping status: 200
	// Storage is available: ✓
}
