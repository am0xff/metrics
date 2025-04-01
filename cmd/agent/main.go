package main

import (
	"time"

	"github.com/am0xff/metrics/internal/reporter"
)

func main() {
	opt := parseFlags()

	pollInterval := time.Duration(opt.pollInterval) * time.Second
	reportInterval := time.Duration(opt.reportInterval) * time.Second

	// Инициализируем resty-клиент с заданным адресом сервера.
	reporter.InitClient(opt.addr)

	// Создаем экземпляр метрик.
	metrics := reporter.NewMetrics()

	go func() {
		for {
			reporter.UpdateMetrics(metrics)
			time.Sleep(pollInterval)
		}
	}()

	go func() {
		for {
			reporter.ReportMetrics(metrics)
			time.Sleep(reportInterval)
		}
	}()

	select {}
}
