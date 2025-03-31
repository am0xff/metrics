package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

const (
	APIUrl         = "http://localhost:8080"
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	contentType    = "text/plain"
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

func updateMetrics(m *Metrics) {
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

func reportMetrics(m *Metrics) {
	// Отправка gauges метрик
	for name, value := range m.Gauges {
		url := fmt.Sprintf("%s/update/gauge/%s/%v", APIUrl, name, value)
		req, err := http.Post(url, contentType, nil)
		if err != nil {
			log.Printf("%s %s: %v", err, name, err)
			continue
		}

		io.Copy(io.Discard, req.Body)
		req.Body.Close()
	}

	// Отправка counter метрик
	for name, value := range m.Counters {
		url := fmt.Sprintf("%s/update/counter/%s/%d", APIUrl, name, value)
		req, err := http.Post(url, contentType, nil)
		if err != nil {
			log.Printf("%s %s: %v", err, name, err)
			continue
		}

		io.Copy(io.Discard, req.Body)
		req.Body.Close()
	}
}

func main() {
	metrics := NewMetrics()

	go func() {
		for {
			updateMetrics(metrics)
			time.Sleep(pollInterval) // pollInterval = 2 * time.Second
		}
	}()

	go func() {
		for {
			reportMetrics(metrics)
			time.Sleep(reportInterval) // reportInterval = 10 * time.Second
		}
	}()

	select {}
}
