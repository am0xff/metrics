package agent

import (
	"fmt"
	"log"
	"time"
)

type Agent struct {
	config    Config
	collector *Collector
	reporter  *Reporter
	gauges    map[string]float64
	counters  map[string]int64
}

func NewAgent(config Config) *Agent {
	return &Agent{
		config:    config,
		collector: NewCollector(),
		reporter:  NewReporter(config.ServerAddr),
		gauges:    make(map[string]float64),
		counters:  make(map[string]int64),
	}
}

func Run() error {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	agent := NewAgent(cfg)
	fmt.Println("Running agent on", cfg.ServerAddr)

	go func() {
		for {
			g, c := agent.collector.Collect()
			agent.gauges = g
			agent.counters = c
			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
		}
	}()

	go func() {
		for {
			g := agent.gauges
			c := agent.counters
			if g != nil && c != nil {
				agent.reporter.ReportBatch(g, c)
			}
			time.Sleep(time.Duration(cfg.ReportInterval) * time.Second)
		}
	}()

	select {}
}
