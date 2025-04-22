package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/am0xff/metrics/internal/models"
	"github.com/am0xff/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	storage *storage.MemStorage
}

func NewHandler(storage *storage.MemStorage) *Handler {
	return &Handler{storage: storage}
}

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

func (h *Handler) GETGetMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch metricType {
	case "gauge":
		v, ok := h.storage.Gauges.Get(name)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, strconv.FormatFloat(float64(v), 'f', -1, 64))
	case "counter":
		v, ok := h.storage.Counters.Get(name)
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

func (h *Handler) GETUpdateMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	valueStr := chi.URLParam(r, "value")

	if name == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.storage.Gauges.Set(name, storage.Gauge(value))
	case "counter":
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		h.storage.Counters.Set(name, storage.Counter(value))
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
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
