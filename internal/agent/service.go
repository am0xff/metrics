package agent

import (
	"fmt"
	"github.com/am0xff/metrics/internal/models"
	"github.com/am0xff/metrics/internal/storage"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"log"
	"time"
)

type Agent struct {
	cfg      Config
	reporter *Reporter
	jobs     chan models.Metrics
}

func NewAgent(cfg Config) *Agent {
	return &Agent{
		cfg: cfg,
		reporter: NewReporter(&ReporterConfig{
			ServerAddr: cfg.ServerAddr,
			Key:        cfg.Key,
		}),
		jobs: make(chan models.Metrics, cfg.RateLimit),
	}
}

func Run() error {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	agent := NewAgent(cfg)
	fmt.Println("Running agent on", cfg.ServerAddr)

	for i := 0; i < cfg.RateLimit; i++ {
		go func() {
			for m := range agent.jobs {
				agent.reporter.send(m.MType, m.ID, m.String())
			}
		}()
	}

	go func() {
		collector := NewCollector()
		ticker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			g, c := collector.Collect()
			for name, v := range g {
				agent.jobs <- models.Metrics{ID: name, MType: storage.MetricTypeGauge, Value: &v}
			}
			for name, d := range c {
				agent.jobs <- models.Metrics{ID: name, MType: storage.MetricTypeCounter, Delta: &d}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			vm, _ := mem.VirtualMemory()

			total := float64(vm.Total)
			free := float64(vm.Free)
			agent.jobs <- models.Metrics{ID: "TotalMemory", MType: storage.MetricTypeGauge, Value: &total}
			agent.jobs <- models.Metrics{ID: "FreeMemory", MType: storage.MetricTypeGauge, Value: &free}

			percents, _ := cpu.Percent(1*time.Second, true)
			for idx, pct := range percents {
				name := fmt.Sprintf("CPUutilization%d", idx+1)
				p := pct
				agent.jobs <- models.Metrics{
					ID:    name,
					MType: storage.MetricTypeGauge,
					Value: &p,
				}
			}
		}
	}()

	select {}
}
