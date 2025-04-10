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

type Collector struct {
	pollCount int64
}

func NewCollector() *Collector {
	return &Collector{pollCount: 0}
}

func (c *Collector) Collect() (map[string]float64, map[string]int64) {
	gauges := make(map[string]float64)
	counters := make(map[string]int64)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	gauges["Alloc"] = float64(m.Alloc)
	gauges["BuckHashSys"] = float64(m.BuckHashSys)
	gauges["Frees"] = float64(m.Frees)
	gauges["GCCPUFraction"] = m.GCCPUFraction
	gauges["GCSys"] = float64(m.GCSys)
	gauges["HeapAlloc"] = float64(m.HeapAlloc)
	gauges["HeapIdle"] = float64(m.HeapIdle)
	gauges["HeapInuse"] = float64(m.HeapInuse)
	gauges["HeapObjects"] = float64(m.HeapObjects)
	gauges["HeapReleased"] = float64(m.HeapReleased)
	gauges["HeapSys"] = float64(m.HeapSys)
	gauges["LastGC"] = float64(m.LastGC)
	gauges["Lookups"] = float64(m.Lookups)
	gauges["MCacheInuse"] = float64(m.MCacheInuse)
	gauges["MCacheSys"] = float64(m.MCacheSys)
	gauges["MSpanInuse"] = float64(m.MSpanInuse)
	gauges["MSpanSys"] = float64(m.MSpanSys)
	gauges["Mallocs"] = float64(m.Mallocs)
	gauges["NextGC"] = float64(m.NextGC)
	gauges["NumForcedGC"] = float64(m.NumForcedGC)
	gauges["NumGC"] = float64(m.NumGC)
	gauges["OtherSys"] = float64(m.OtherSys)
	gauges["PauseTotalNs"] = float64(m.PauseTotalNs)
	gauges["StackInuse"] = float64(m.StackInuse)
	gauges["StackSys"] = float64(m.StackSys)
	gauges["Sys"] = float64(m.Sys)
	gauges["TotalAlloc"] = float64(m.TotalAlloc)
	gauges["RandomValue"] = rand.Float64()

	c.pollCount++
	counters["PollCount"] = c.pollCount

	return gauges, counters
}

type Reporter struct {
	serverAddr string
}

func NewReporter(serverAddr string) *Reporter {
	return &Reporter{serverAddr}
}

func (r *Reporter) reportMetric(metricType, name, valueStr string) {
	url := fmt.Sprintf("http://%s/update/%s/%s/%s", r.serverAddr, metricType, name, valueStr)
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

func (r *Reporter) Report(gauges map[string]float64, counters map[string]int64) {
	for name, value := range gauges {
		valueStr := strconv.FormatFloat(value, 'f', -1, 64)
		r.reportMetric("gauge", name, valueStr)
	}
	for name, value := range counters {
		valueStr := strconv.FormatInt(value, 10)
		r.reportMetric("counter", name, valueStr)
	}
}

type Agent struct {
	config    Config
	collector *Collector
	reporter  *Reporter
}

func NewAgent(config Config) *Agent {
	return &Agent{
		config:    config,
		collector: NewCollector(),
		reporter:  NewReporter(config.ServerAddr),
	}
}

func Run() {
	var config Config

	if err := env.Parse(&config); err != nil {
		log.Fatalf("Parse env: %v", err)
	}

	serverAddr := flag.String("a", config.ServerAddr, "HTTP сервер адрес")
	reportInterval := flag.Int("r", config.ReportInterval, "Интервал отправки метрик (сек)")
	pollInterval := flag.Int("p", config.PollInterval, "Интервал опроса метрик (сек)")
	flag.Parse()

	config = Config{
		ServerAddr:     *serverAddr,
		ReportInterval: *reportInterval,
		PollInterval:   *pollInterval,
	}

	agent := NewAgent(config)
	fmt.Println("Running agent on", config.ServerAddr)

	var latestGauges map[string]float64
	var latestCounters map[string]int64

	go func() {
		for {
			g, c := agent.collector.Collect()
			latestGauges = g
			latestCounters = c
			time.Sleep(time.Duration(config.PollInterval) * time.Second)
		}
	}()

	go func() {
		for {
			if latestGauges != nil && latestCounters != nil {
				agent.reporter.Report(latestGauges, latestCounters)
			}
			time.Sleep(time.Duration(config.ReportInterval) * time.Second)
		}
	}()

	select {}
}
