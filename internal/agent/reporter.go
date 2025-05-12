package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/am0xff/metrics/internal/models"
	"github.com/am0xff/metrics/internal/storage"
	"log"
	"net/http"
	"strconv"
)

// Reporter отправляет метрики на HTTP‑сервер.
type Reporter struct {
	serverAddr string
	client     *http.Client
}

// NewReporter создаёт отправщика.
func NewReporter(serverAddr string) *Reporter {
	return &Reporter{
		serverAddr: serverAddr,
		client:     http.DefaultClient,
	}
}

func (r *Reporter) Report(gauges map[string]float64, counters map[string]int64) {
	for name, v := range gauges {
		r.send(storage.MetricTypeGauge, name, strconv.FormatFloat(v, 'f', -1, 64))
	}
	for name, v := range counters {
		r.send(storage.MetricTypeCounter, name, strconv.FormatInt(v, 10))
	}
}

func (r *Reporter) ReportBatch(gauges map[string]float64, counters map[string]int64) {
	r.sendBatch(gauges, counters)
}

func (r *Reporter) send(metricType storage.MetricType, name, value string) {
	m := models.Metrics{
		ID:    name,
		MType: metricType,
	}

	switch metricType {
	case storage.MetricTypeGauge:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Println("invalid gauge value:", err)
			return
		}
		m.Value = &v
	case storage.MetricTypeCounter:
		d, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			log.Println("invalid counter value:", err)
			return
		}
		m.Delta = &d
	default:
		log.Printf("unsupported metric type: %s\n", metricType)
		return
	}

	data, err := json.Marshal(m)
	if err != nil {
		log.Println("json marshal failed:", err)
		return
	}

	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		log.Println("gzip compression failed:", err)
		return
	}

	if _, err := gz.Write(data); err != nil {
		log.Println("gzip write failed:", err)
		return
	}
	if err := gz.Close(); err != nil {
		log.Println("gzip close failed:", err)
		return
	}

	url := fmt.Sprintf("http://%s/update/", r.serverAddr)
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		log.Println("create request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := r.client.Do(req)
	if err != nil {
		log.Println("send metric:", err)
		return
	}

	if err := resp.Body.Close(); err != nil {
		log.Println("close response body:", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("bad status %d for %s/%s\n", resp.StatusCode, metricType, name)
	}
}

func (r *Reporter) sendBatch(gauges map[string]float64, counters map[string]int64) {
	var metrics []models.Metrics

	for name, v := range gauges {
		value, err := strconv.ParseFloat(strconv.FormatFloat(v, 'f', -1, 64), 64)
		if err != nil {
			log.Println("invalid gauge value:", err)
			return
		}
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: storage.MetricTypeGauge,
			Value: &value,
		})
	}

	for name, d := range counters {
		delta, err := strconv.ParseInt(strconv.FormatInt(d, 10), 10, 64)
		if err != nil {
			log.Println("invalid counter value:", err)
			return
		}
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: storage.MetricTypeCounter,
			Delta: &delta,
		})
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		log.Println("json marshal failed:", err)
		return
	}

	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		log.Println("gzip compression failed:", err)
		return
	}

	if _, err := gz.Write(data); err != nil {
		log.Println("gzip write failed:", err)
		return
	}
	if err := gz.Close(); err != nil {
		log.Println("gzip close failed:", err)
		return
	}

	url := fmt.Sprintf("http://%s/updates/", r.serverAddr)
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		log.Println("create request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := r.client.Do(req)
	if err != nil {
		log.Println("send metric:", err)
		return
	}

	if err := resp.Body.Close(); err != nil {
		log.Println("close response body:", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("bad status %d when batch", resp.StatusCode)
	}
}
