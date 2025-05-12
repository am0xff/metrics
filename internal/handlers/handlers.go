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
	storageProvider storage.StorageProvider
}

func NewHandler(sp storage.StorageProvider) *Handler {
	return &Handler{storageProvider: sp}
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

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.storageProvider.Ping(r.Context()); err != nil {
		http.Error(w, "database ping failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
