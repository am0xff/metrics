package main

import (
	"net/http"
	"strconv"
	"strings"
)

type gauge float64
type counter int64

type MemStorage struct {
	storageGauge   map[string]gauge
	storageCounter map[string]counter
}

func NewMemStorage() *MemStorage {
	return &MemStorage{storageGauge: make(map[string]gauge), storageCounter: make(map[string]counter)}
}

func (m *MemStorage) SetGauge(name string, value gauge) {
	m.storageGauge[name] = value
}

func (m *MemStorage) SetCounter(name string, value counter) {
	m.storageCounter[name] += value
}

var storage = NewMemStorage()

func apiGauge(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что метод POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Content-Type
	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
	}

	parts := strings.Split(r.URL.Path, "/")
	// parts[0] пустая, далее "update", "gauge", "<name>", "<value>"
	if len(parts) != 5 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metricName := parts[3]
	metricValueStr := parts[4]

	// Если имя метрики отсутствует
	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Преобразуем значение метрики в float64 для gauge
	value, err := strconv.ParseFloat(metricValueStr, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	storage.SetGauge(metricName, gauge(value))
	w.WriteHeader(http.StatusOK)
}

func apiCounter(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что метод POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Content-Type
	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metricName := parts[3]
	metricValueStr := parts[4]

	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Преобразуем значение метрики в int64 для counter
	value, err := strconv.ParseInt(metricValueStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	storage.SetCounter(metricName, counter(value))
	w.WriteHeader(http.StatusOK)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/update/gauge/", apiGauge)
	mux.HandleFunc("/update/counter/", apiCounter)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
