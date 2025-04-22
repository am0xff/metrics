package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/am0xff/metrics/internal/models"
	"github.com/am0xff/metrics/internal/storage"
	"io"
	"net/http"
	"strings"
)

type Handler struct {
	storage *storage.MemStorage
}

func NewHandler(storage *storage.MemStorage) *Handler {
	return &Handler{storage: storage}
}

func (h *Handler) GetMetric(w http.ResponseWriter, r *http.Request) {
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
	case "gauge":
		v, ok := h.storage.Gauges.Get(req.ID)
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
	case "counter":
		v, ok := h.storage.Counters.Get(req.ID)
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

func (h *Handler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
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
	case "gauge":
		if req.Value == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		newValue := storage.Gauge(*req.Value)
		h.storage.Gauges.Set(req.ID, newValue)

		resp = models.Metrics{
			ID:    req.ID,
			MType: req.MType,
			Value: req.Value,
		}
	case "counter":
		if req.Delta == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		newValue := storage.Counter(*req.Delta)
		h.storage.Counters.Set(req.ID, newValue)

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

func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var html strings.Builder

	html.WriteString("<html><head><title>Metrics</title></head><body>")
	html.WriteString("<ul>")
	for _, k := range h.storage.Gauges.Keys() {
		v, _ := h.storage.Gauges.Get(k)
		html.WriteString(fmt.Sprintf("<li>%s: %v</li>", k, v))
	}
	for _, k := range h.storage.Counters.Keys() {
		v, _ := h.storage.Counters.Get(k)
		html.WriteString(fmt.Sprintf("<li>%s: %v</li>", k, v))
	}
	html.WriteString("</ul>")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if _, err := io.WriteString(w, html.String()); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}

//Сервер должен быть доступен по адресу `http://localhost:8080`, а также:
//
//- Принимать и хранить произвольные метрики двух типов:
//- Тип `gauge`, `float64` — новое значение должно замещать предыдущее.
//- Тип `counter`, `int64` — новое значение должно добавляться к предыдущему, если какое-то значение уже было известно серверу.
//- Принимать метрики по протоколу HTTP методом `POST`.
//- Принимать данные в формате
// `http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>`,
// `Content-Type: text/plain`.
//- При успешном приёме возвращать `http.StatusOK`.
//- При попытке передать запрос без имени метрики возвращать `http.StatusNotFound`.
//- При попытке передать запрос с некорректным типом метрики или значением возвращать
//  `http.StatusBadRequest`.

// Доработайте сервер так, чтобы в ответ на запрос
// `GET http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>` он возвращал аккумулированное значение метрики в текстовом виде со статусом `http.StatusOK`.
// При попытке запроса неизвестной метрики сервер должен возвращать
// `http.StatusNotFound`.
// По запросу `GET http://<АДРЕС_СЕРВЕРА>/`
// сервер должен отдавать HTML-страницу со списком имён и значений всех известных ему на текущий момент метрик.

// Для передачи метрик на сервер используйте Content-Type: application/json.
// В теле запроса должен быть описанный выше JSON.
// Передавать метрики нужно через POST update/.
// В теле ответа отправляйте JSON той же структуры с актуальным (изменённым) значением Value.

// SET
// curl -X POST http://localhost:8080/update/ \
//  -H "Content-Type: application/json" \
//  -d '{"id":"someMetric","type":"gauge","value":42.7}'
//curl -X POST http://localhost:8080/update/ \
//  -H "Content-Type: application/json" \
//  -d '{"id":"requests","type":"counter","delta":100}'

// GET
//curl -X POST http://localhost:8080/value/ \
//  -H "Content-Type: application/json" \
//  -d '{"id":"someMetric","type":"gauge"}'
//curl -X POST http://localhost:8080/value/ \
//-H "Content-Type: application/json" \
//-d '{"id":"requests","type":"counter"}'
