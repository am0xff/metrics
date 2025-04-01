package reporter

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"runtime"

	"github.com/go-resty/resty/v2"
)

var (
	client      *resty.Client
	APIUrl      string
	contentType = "text/plain"
)

type Metrics struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewMetrics() *Metrics {
	return &Metrics{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

func InitClient(apiUrl string) {
	APIUrl = apiUrl
	client = resty.New()
}

func UpdateMetrics(m *Metrics) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.Gauges["Alloc"] = float64(memStats.Alloc)
	m.Gauges["BuckHashSys"] = float64(memStats.BuckHashSys)
	m.Gauges["Frees"] = float64(memStats.Frees)
	m.Gauges["GCCPUFraction"] = float64(memStats.GCCPUFraction)
	m.Gauges["GCSys"] = float64(memStats.GCSys)
	m.Gauges["HeapAlloc"] = float64(memStats.HeapAlloc)
	m.Gauges["HeapIdle"] = float64(memStats.HeapIdle)
	m.Gauges["HeapInuse"] = float64(memStats.HeapInuse)
	m.Gauges["HeapObjects"] = float64(memStats.HeapObjects)
	m.Gauges["HeapReleased"] = float64(memStats.HeapReleased)
	m.Gauges["HeapSys"] = float64(memStats.HeapSys)
	m.Gauges["LastGC"] = float64(memStats.LastGC)
	m.Gauges["Lookups"] = float64(memStats.Lookups)
	m.Gauges["MCacheInuse"] = float64(memStats.MCacheInuse)
	m.Gauges["MCacheSys"] = float64(memStats.MCacheSys)
	m.Gauges["MSpanInuse"] = float64(memStats.MSpanInuse)
	m.Gauges["MSpanSys"] = float64(memStats.MSpanSys)
	m.Gauges["Mallocs"] = float64(memStats.Mallocs)
	m.Gauges["NextGC"] = float64(memStats.NextGC)
	m.Gauges["NumForcedGC"] = float64(memStats.NumForcedGC)
	m.Gauges["NumGC"] = float64(memStats.NumGC)
	m.Gauges["OtherSys"] = float64(memStats.OtherSys)
	m.Gauges["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	m.Gauges["StackInuse"] = float64(memStats.StackInuse)
	m.Gauges["StackSys"] = float64(memStats.StackSys)
	m.Gauges["Sys"] = float64(memStats.Sys)
	m.Gauges["TotalAlloc"] = float64(memStats.TotalAlloc)
	m.Gauges["RandomValue"] = rand.Float64()

	m.Counters["PollCount"]++
}

func sendMetric(metricType, name string, value interface{}) error {
	var url string
	switch metricType {
	case "gauge":
		url = fmt.Sprintf("%s/update/gauge/%s/%v", APIUrl, name, value)
	case "counter":
		if v, ok := value.(int64); ok {
			url = fmt.Sprintf("%s/update/counter/%s/%d", APIUrl, name, v)
		} else {
			return fmt.Errorf("value for counter %s is not int64", name)
		}
	default:
		return fmt.Errorf("unknown metric type: %s", metricType)
	}

	resp, err := client.R().
		SetHeader("Content-Type", contentType).
		Post(url)
	if err != nil {
		return err
	}

	_, err = io.Copy(io.Discard, resp.RawBody())
	if err != nil {
		return err
	}
	resp.RawBody().Close()
	return nil
}

func ReportMetrics(m *Metrics) {
	for name, value := range m.Gauges {
		if err := sendMetric("gauge", name, value); err != nil {
			log.Printf("Error sending gauge %s: %v", name, err)
		}
	}
	for name, value := range m.Counters {
		if err := sendMetric("counter", name, value); err != nil {
			log.Printf("Error sending counter %s: %v", name, err)
		}
	}
}
