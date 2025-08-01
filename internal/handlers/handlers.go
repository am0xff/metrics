// Package handlers предоставляет HTTP обработчики для API сервиса метрик.
// Пакет содержит обработчики для операций получения, обновления и просмотра метрик
// через REST API эндпоинты. Поддерживает работу с метриками типов gauge и counter.
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/am0xff/metrics/internal/models"
	"github.com/am0xff/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

// Handler представляет структуру для обработки HTTP запросов к API метрик.
// Содержит ссылку на провайдер хранилища для выполнения операций с данными.
//
// Пример использования:
//
//	storage := memory.NewMemoryStorage()
//	handler := NewHandler(storage)
//	http.HandleFunc("/metrics", handler.GetMetrics)
type Handler struct {
	storageProvider storage.StorageProvider
}

// NewHandler создает новый экземпляр Handler с указанным провайдером хранилища.
// Провайдер хранилища должен реализовывать интерфейс storage.StorageProvider.
//
// Пример использования:
//
//	memStorage := memory.NewMemoryStorage()
//	handler := NewHandler(memStorage)
//
//	// Использование с HTTP сервером
//	mux := http.NewServeMux()
//	mux.HandleFunc("/metrics", handler.GetMetrics)
func NewHandler(sp storage.StorageProvider) *Handler {
	return &Handler{storageProvider: sp}
}

// POSTGetMetric обрабатывает POST запросы для получения значения метрики в формате JSON.
// Принимает JSON с указанием типа и имени метрики, возвращает её текущее значение.
//
// Ожидаемый формат запроса:
//
//	{
//		"id": "metric_name",
//		"type": "gauge" | "counter"
//	}
//
// Формат ответа для gauge:
//
//	{
//		"id": "metric_name",
//		"type": "gauge",
//		"value": 123.45
//	}
//
// Формат ответа для counter:
//
//	{
//		"id": "metric_name",
//		"type": "counter",
//		"delta": 100
//	}
//
// HTTP статусы:
//   - 200: метрика найдена и возвращена
//   - 400: неверный формат запроса или тип метрики
//   - 404: метрика не найдена или отсутствуют обязательные поля
//   - 405: неверный HTTP метод (ожидается POST)
func (h *Handler) POSTGetMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.MType == "" || req.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var resp models.Metrics

	switch req.MType {
	case storage.MetricTypeGauge:
		v, ok := h.storageProvider.GetGauge(r.Context(), req.ID)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		value := float64(v)
		resp = models.Metrics{
			ID:    req.ID,
			MType: req.MType,
			Value: &value,
		}
	case storage.MetricTypeCounter:
		v, ok := h.storageProvider.GetCounter(r.Context(), req.ID)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		value := int64(v)
		resp = models.Metrics{
			ID:    req.ID,
			MType: req.MType,
			Delta: &value,
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		return
	}
}

// POSTUpdateMetric обрабатывает POST запросы для обновления значения метрики в формате JSON.
// Принимает JSON с типом, именем и новым значением метрики.
//
// Ожидаемый формат запроса для gauge:
//
//	{
//		"id": "metric_name",
//		"type": "gauge",
//		"value": 123.45
//	}
//
// Ожидаемый формат запроса для counter:
//
//	{
//		"id": "metric_name",
//		"type": "counter",
//		"delta": 10
//	}
//
// Возвращает обновленную метрику в том же формате.
//
// HTTP статусы:
//   - 200: метрика успешно обновлена
//   - 400: неверный формат запроса, тип метрики или отсутствует значение
//   - 404: отсутствуют обязательные поля (id или type)
//   - 405: неверный HTTP метод (ожидается POST)
func (h *Handler) POSTUpdateMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.MType == "" || req.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var resp models.Metrics

	switch req.MType {
	case storage.MetricTypeGauge:
		if req.Value == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		newValue := storage.Gauge(*req.Value)
		h.storageProvider.SetGauge(r.Context(), req.ID, newValue)

		resp = models.Metrics{
			ID:    req.ID,
			MType: req.MType,
			Value: req.Value,
		}
	case storage.MetricTypeCounter:
		if req.Delta == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		newValue := storage.Counter(*req.Delta)
		h.storageProvider.SetCounter(r.Context(), req.ID, newValue)

		resp = models.Metrics{
			ID:    req.ID,
			MType: req.MType,
			Delta: req.Delta,
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		return
	}
}

// POSTUpdatesMetrics обрабатывает POST запросы для массового обновления метрик в формате JSON.
// Принимает массив метрик и выполняет пакетное обновление.
//
// Ожидаемый формат запроса:
//
//	[
//		{
//			"id": "cpu_usage",
//			"type": "gauge",
//			"value": 85.5
//		},
//		{
//			"id": "requests_total",
//			"type": "counter",
//			"delta": 100
//		}
//	]
//
// HTTP статусы:
//   - 200: все метрики успешно обновлены
//   - 400: неверный формат запроса, пустой массив или неверные данные метрики
//   - 404: отсутствуют обязательные поля в одной из метрик
//   - 405: неверный HTTP метод (ожидается POST)
func (h *Handler) POSTUpdatesMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var reqs []models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&reqs); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(reqs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, req := range reqs {
		if req.MType == "" || req.ID == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch req.MType {
		case storage.MetricTypeGauge:
			if req.Value == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			newValue := storage.Gauge(*req.Value)
			h.storageProvider.SetGauge(r.Context(), req.ID, newValue)
		case storage.MetricTypeCounter:
			if req.Delta == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			newValue := storage.Counter(*req.Delta)
			h.storageProvider.SetCounter(r.Context(), req.ID, newValue)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// GETGetMetric обрабатывает GET запросы для получения значения метрики через URL параметры.
// Возвращает значение метрики в текстовом формате.
//
// URL формат: /value/{type}/{name}
// где:
//   - type: "gauge" или "counter"
//   - name: имя метрики
//
// Примеры URL:
//   - /value/gauge/cpu_usage
//   - /value/counter/requests_total
//
// HTTP статусы:
//   - 200: метрика найдена, значение возвращено в теле ответа
//   - 400: неверный тип метрики
//   - 404: метрика не найдена
func (h *Handler) GETGetMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch storage.MetricType(metricType) {
	case storage.MetricTypeGauge:
		v, ok := h.storageProvider.GetGauge(r.Context(), name)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, strconv.FormatFloat(float64(v), 'f', -1, 64))
	case storage.MetricTypeCounter:
		v, ok := h.storageProvider.GetCounter(r.Context(), name)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, strconv.FormatInt(int64(v), 10))
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GETUpdateMetric обрабатывает POST запросы для обновления метрики через URL параметры.
// Обновляет значение метрики, переданное в URL.
//
// URL формат: /update/{type}/{name}/{value}
// где:
//   - type: "gauge" или "counter"
//   - name: имя метрики
//   - value: новое значение метрики
//
// Примеры URL:
//   - /update/gauge/cpu_usage/85.5
//   - /update/counter/requests_total/1000
//
// HTTP статусы:
//   - 200: метрика успешно обновлена
//   - 400: неверный тип метрики или формат значения
//   - 404: не указано имя метрики
func (h *Handler) GETUpdateMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	valueStr := chi.URLParam(r, "value")

	if name == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch storage.MetricType(metricType) {
	case storage.MetricTypeGauge:
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.storageProvider.SetGauge(r.Context(), name, storage.Gauge(value))
	case storage.MetricTypeCounter:
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		h.storageProvider.SetCounter(r.Context(), name, storage.Counter(value))
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetMetrics обрабатывает GET запросы для получения списка всех метрик в HTML формате.
// Возвращает HTML страницу со списком всех gauge и counter метрик с их значениями.
//
// URL: /
//
// Формат ответа: HTML страница с неупорядоченным списком метрик.
// Каждая метрика отображается в формате "имя: значение".
//
// HTTP статусы:
//   - 200: страница с метриками успешно возвращена
//   - 405: неверный HTTP метод (ожидается GET)
//   - 500: ошибка при формировании ответа
func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var html strings.Builder

	html.WriteString("<html><head><title>Metrics</title></head><body>")
	html.WriteString("<ul>")
	for _, k := range h.storageProvider.KeysGauge(r.Context()) {
		v, _ := h.storageProvider.GetGauge(r.Context(), k)
		html.WriteString(fmt.Sprintf("<li>%s: %v</li>", k, v))
	}
	for _, k := range h.storageProvider.KeysCounter(r.Context()) {
		v, _ := h.storageProvider.GetCounter(r.Context(), k)
		html.WriteString(fmt.Sprintf("<li>%s: %v</li>", k, v))
	}
	html.WriteString("</ul>")

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	if _, err := io.WriteString(w, html.String()); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}

// Ping обрабатывает запросы для проверки доступности хранилища.
// Выполняет проверку соединения с базой данных или другим хранилищем.
//
// URL: /ping
//
// HTTP статусы:
//   - 200: хранилище доступно
//   - 500: хранилище недоступно или ошибка соединения
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.storageProvider.Ping(r.Context()); err != nil {
		http.Error(w, "database ping failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
