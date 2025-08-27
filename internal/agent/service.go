package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/am0xff/metrics/internal/models"
	"github.com/am0xff/metrics/internal/storage"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
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
			CryptoKey:  cfg.CryptoKey,
		}),
		jobs: make(chan models.Metrics, cfg.RateLimit),
	}
}

func Run() error {
	var wg sync.WaitGroup

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	agent := NewAgent(cfg)
	fmt.Println("Running agent on", cfg.ServerAddr)

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal: %v. Shutting down gracefully...\n", sig)
		cancel()
	}()

	for i := 0; i < cfg.RateLimit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case m := <-agent.jobs:
					agent.reporter.send(m.MType, m.ID, m.String())
				}
			}
		}()
	}

	wg.Add(1)
	go func() {
		collector := NewCollector()
		ticker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
		defer ticker.Stop()
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				g, c := collector.Collect()

				for name, v := range g {
					select {
					case <-ctx.Done():
						return
					case agent.jobs <- models.Metrics{ID: name, MType: storage.MetricTypeGauge, Value: &v}:
					}
				}
				for name, d := range c {
					select {
					case <-ctx.Done():
						return
					case agent.jobs <- models.Metrics{ID: name, MType: storage.MetricTypeCounter, Delta: &d}:
					}
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
		defer ticker.Stop()
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				vm, _ := mem.VirtualMemory()
				total := float64(vm.Total)
				free := float64(vm.Free)

				select {
				case <-ctx.Done():
					return
				case agent.jobs <- models.Metrics{ID: "TotalMemory", MType: storage.MetricTypeGauge, Value: &total}:
				}

				select {
				case <-ctx.Done():
					return
				case agent.jobs <- models.Metrics{ID: "FreeMemory", MType: storage.MetricTypeGauge, Value: &free}:
				}

				percents, _ := cpu.Percent(1*time.Second, true)
				for idx, pct := range percents {
					name := fmt.Sprintf("CPUutilization%d", idx+1)
					p := pct
					select {
					case <-ctx.Done():
						return
					case agent.jobs <- models.Metrics{
						ID:    name,
						MType: storage.MetricTypeGauge,
						Value: &p,
					}:
					}
				}
			}
		}
	}()

	wg.Wait()
	close(agent.jobs)

	return nil
}
