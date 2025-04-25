package agent

import (
	"math/rand"
	"runtime"
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
