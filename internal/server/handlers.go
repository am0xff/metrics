package server

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// UpdateMetricHandler обрабатывает POST запросы вида:
// /update/{type}/{name}/{value}
func (s *Server) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	valueStr := chi.URLParam(r, "value")

	if name == "" {
		http.Error(w, "Metric name not provided", http.StatusNotFound)
		return
	}

	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		if err := s.Storage.UpdateGauge(name, value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "counter":
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		if err := s.Storage.UpdateCounter(name, value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "OK")
}

// GetMetricHandler обрабатывает GET запросы вида:
// /value/{type}/{name}
func (s *Server) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch metricType {
	case "gauge":
		value, ok := s.Storage.GetGauge(name)
		if !ok {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, strconv.FormatFloat(value, 'f', -1, 64))
	case "counter":
		value, ok := s.Storage.GetCounter(name)
		if !ok {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, strconv.FormatInt(value, 10))
	default:
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
		return
	}
}

// ListMetricsHandler отдает HTML со списком всех метрик.
func (s *Server) ListMetricsHandler(rw http.ResponseWriter, r *http.Request) {
	html := "<html><head><title>Metrics</title></head><body>"
	html += "<ul>"
	for name, value := range s.Storage.GetAllGauges() {
		html += fmt.Sprintf("<li>%s: %v</li>", name, value)
	}
	for name, value := range s.Storage.GetAllCounters() {
		html += fmt.Sprintf("<li>%s: %v</li>", name, value)
	}
	html += "</ul>"

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(http.StatusOK)

	rw.Write([]byte(html))
}
