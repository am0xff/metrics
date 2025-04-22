package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/am0xff/metrics/internal/models"
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
		r.send("gauge", name, strconv.FormatFloat(v, 'f', -1, 64))
	}
	for name, v := range counters {
		r.send("counter", name, strconv.FormatInt(v, 10))
	}
}

func (r *Reporter) send(metricType, name, value string) {
	m := models.Metrics{
		ID:    name,
		MType: metricType,
	}

	switch metricType {
	case "gauge":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Println("invalid gauge value:", err)
			return
		}
		m.Value = &v
	case "counter":
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

	payload, err := json.Marshal(m)
	if err != nil {
		log.Println("json marshal failed:", err)
		return
	}

	url := fmt.Sprintf("http://%s/update/", r.serverAddr)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		log.Println("create request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Отправляем
	resp, err := r.client.Do(req)
	if err != nil {
		log.Println("send metric:", err)
		return
	}
	defer resp.Body.Close()

	// Проверяем ответ
	if resp.StatusCode != http.StatusOK {
		log.Printf("bad status %d for %s/%s\n", resp.StatusCode, metricType, name)
	}
}
