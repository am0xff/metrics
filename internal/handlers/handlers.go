package handlers

import (
	"github.com/am0xff/metrics/internal/storage"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	Storage *storage.MemStorage
}

func NewHandler(s *storage.MemStorage) *Handler {
	return &Handler{Storage: s}
}

func (h *Handler) APIUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	parts := strings.Split(r.URL.Path, "/")
	// parts[0] пустая, далее "update", "<тип>", "<имя>", "<значение>"
	if len(parts) != 5 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metricType := parts[2]
	metricName := parts[3]
	metricValueStr := parts[4]

	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValueStr, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.Storage.SetGauge(metricName, value)
		w.WriteHeader(http.StatusOK)
	case "counter":
		value, err := strconv.ParseInt(metricValueStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.Storage.SetCounter(metricName, value)
		w.WriteHeader(http.StatusOK)
	default:
		// Для неизвестного типа возвращаем 404
		w.WriteHeader(http.StatusNotFound)
	}
}
