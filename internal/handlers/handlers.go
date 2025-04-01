package handlers

import (
	"fmt"
	"github.com/am0xff/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type Handler struct {
	Storage *storage.MemStorage
}

func NewHandler(s *storage.MemStorage) *Handler {
	return &Handler{Storage: s}
}

func (h *Handler) RootHandle(rw http.ResponseWriter, r *http.Request) {
	html := "<html><head><title>Metrics</title></head><body>"
	html += "<ul>"
	for name, value := range h.Storage.GaugeValues() {
		html += fmt.Sprintf("<li>%s: %v</li>", name, value)
	}
	for name, value := range h.Storage.CounterValues() {
		html += fmt.Sprintf("<li>%s: %v</li>", name, value)
	}
	html += "</ul>"

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(http.StatusOK)

	rw.Write([]byte(html))
}

func (h *Handler) GetCounterMetric(rw http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metric_name")
	value, ok := h.Storage.GetCounter(metricName)

	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	s := fmt.Sprintf("%v", value)
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(s))
}

func (h *Handler) GetGaugeMetric(rw http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metric_name")
	value, ok := h.Storage.GetGauge(metricName)

	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	s := fmt.Sprintf("%v", value)
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(s))
}

func (h *Handler) UpdateGaugeMetric(rw http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metric_name")
	metricValue := chi.URLParam(r, "metric_value")

	value, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	h.Storage.SetGauge(metricName, value)
	rw.WriteHeader(http.StatusOK)
}
func (h *Handler) UpdateCounterMetric(rw http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metric_name")
	metricValue := chi.URLParam(r, "metric_value")

	value, err := strconv.ParseInt(metricValue, 10, 64)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	h.Storage.SetCounter(metricName, value)
	rw.WriteHeader(http.StatusOK)
}
