package agent

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

type AgentConfig struct {
	ServerAddr     string `env:"ADDRESS" envDefault:"localhost:8080"`
	PollInterval   int    `env:"POLL_INTERVAL" envDefault:"2"`
	ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"10"`
}

type Agent struct {
	config         AgentConfig
	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
}

func NewAgent(config AgentConfig) *Agent {
	return &Agent{
		config:         config,
		gaugeMetrics:   make(map[string]float64),
		counterMetrics: make(map[string]int64),
	}
}

func (a *Agent) collectMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	a.gaugeMetrics["Alloc"] = float64(m.Alloc)
	a.gaugeMetrics["BuckHashSys"] = float64(m.BuckHashSys)
	a.gaugeMetrics["Frees"] = float64(m.Frees)
	a.gaugeMetrics["GCCPUFraction"] = m.GCCPUFraction
	a.gaugeMetrics["GCSys"] = float64(m.GCSys)
	a.gaugeMetrics["HeapAlloc"] = float64(m.HeapAlloc)
	a.gaugeMetrics["HeapIdle"] = float64(m.HeapIdle)
	a.gaugeMetrics["HeapInuse"] = float64(m.HeapInuse)
	a.gaugeMetrics["HeapObjects"] = float64(m.HeapObjects)
	a.gaugeMetrics["HeapReleased"] = float64(m.HeapReleased)
	a.gaugeMetrics["HeapSys"] = float64(m.HeapSys)
	a.gaugeMetrics["LastGC"] = float64(m.LastGC)
	a.gaugeMetrics["Lookups"] = float64(m.Lookups)
	a.gaugeMetrics["MCacheInuse"] = float64(m.MCacheInuse)
	a.gaugeMetrics["MCacheSys"] = float64(m.MCacheSys)
	a.gaugeMetrics["MSpanInuse"] = float64(m.MSpanInuse)
	a.gaugeMetrics["MSpanSys"] = float64(m.MSpanSys)
	a.gaugeMetrics["Mallocs"] = float64(m.Mallocs)
	a.gaugeMetrics["NextGC"] = float64(m.NextGC)
	a.gaugeMetrics["NumForcedGC"] = float64(m.NumForcedGC)
	a.gaugeMetrics["NumGC"] = float64(m.NumGC)
	a.gaugeMetrics["OtherSys"] = float64(m.OtherSys)
	a.gaugeMetrics["PauseTotalNs"] = float64(m.PauseTotalNs)
	a.gaugeMetrics["StackInuse"] = float64(m.StackInuse)
	a.gaugeMetrics["StackSys"] = float64(m.StackSys)
	a.gaugeMetrics["Sys"] = float64(m.Sys)
	a.gaugeMetrics["TotalAlloc"] = float64(m.TotalAlloc)

	a.counterMetrics["PollCount"]++
	a.gaugeMetrics["RandomValue"] = rand.Float64()
}

func (a *Agent) sendMetric(metricType, name, valueStr string) {
	url := fmt.Sprintf("http://%s/update/%s/%s/%s", a.config.ServerAddr, metricType, name, valueStr)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error sending metric:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to send metric %s %s: status %d\n", metricType, name, resp.StatusCode)
	}
}

func (a *Agent) reportMetrics() {
	for name, value := range a.gaugeMetrics {
		a.sendMetric("gauge", name, strconv.FormatFloat(value, 'f', -1, 64))
	}
	for name, value := range a.counterMetrics {
		a.sendMetric("counter", name, strconv.FormatInt(value, 10))
	}
}

func Run() {
	var config AgentConfig

	if err := env.Parse(&config); err != nil {
		log.Fatalf("Parse env: %v", err)
	}

	serverAddr := flag.String("a", config.ServerAddr, "HTTP сервер адрес")
	reportInterval := flag.Int("r", config.ReportInterval, "Интервал отправки метрик (сек)")
	pollInterval := flag.Int("p", config.PollInterval, "Интервал опроса метрик (сек)")
	flag.Parse()

	config = AgentConfig{
		ServerAddr:     *serverAddr,
		ReportInterval: *reportInterval,
		PollInterval:   *pollInterval,
	}

	agent := NewAgent(config)
	log.Println("Запуск агента")

	go func() {
		for {
			agent.collectMetrics()
			time.Sleep(time.Duration(config.PollInterval) * time.Second)
		}
	}()

	go func() {
		for {
			agent.reportMetrics()
			time.Sleep(time.Duration(config.ReportInterval) * time.Second)
		}
	}()

	select {}
}
