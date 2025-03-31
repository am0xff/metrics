package main

import (
	"github.com/am0xff/metrics/internal/reporter"
	"time"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

func main() {
	m := reporter.NewMetrics()

	go func() {
		for {
			reporter.UpdateMetrics(m)
			time.Sleep(pollInterval)
		}
	}()

	go func() {
		for {
			reporter.ReportMetrics(m)
			time.Sleep(reportInterval)
		}
	}()

	select {}
}
