package server

import (
	"encoding/json"
	"fmt"
	"github.com/am0xff/metrics/internal/logger"
	"github.com/am0xff/metrics/internal/models"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

func (s *Server) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
	}

	var req models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch req.MType {
	case "gauge":
		value := *req.Value
		if err := s.Storage.UpdateGauge(req.ID, value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "counter":
		value := *req.Delta
		if err := s.Storage.UpdateCounter(req.ID, value); err != nil {
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

func (s *Server) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
	}

	var req models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch req.MType {
	case "gauge":
		value, ok := s.Storage.GetGauge(req.ID)
		if !ok {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, strconv.FormatFloat(value, 'f', -1, 64))
	case "counter":
		value, ok := s.Storage.GetCounter(req.ID)
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
