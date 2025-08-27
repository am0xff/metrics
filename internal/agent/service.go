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
	cfg          Config
	httpReporter *Reporter
	grpcReporter *GRPCReporterWrapper
	jobs         chan models.Metrics
}

func NewAgent(cfg Config) (*Agent, error) {
	agent := &Agent{
		cfg:  cfg,
		jobs: make(chan models.Metrics, cfg.RateLimit),
	}

	switch cfg.Protocol {
	case "grpc":
		grpcReporter, err := NewGRPCReporter(&ReporterConfig{
			ServerAddr: cfg.GRPCAddr,
			Key:        cfg.Key,
			CryptoKey:  cfg.CryptoKey,
		})
		if err != nil {
			return nil, err
		}
		agent.grpcReporter = grpcReporter

	case "http":
		agent.httpReporter = NewReporter(&ReporterConfig{
			ServerAddr: cfg.ServerAddr,
			Key:        cfg.Key,
			CryptoKey:  cfg.CryptoKey,
		})

	default:
		return nil, fmt.Errorf("unsupported protocol: %s", cfg.Protocol)
	}

	return agent, nil
}

func Run() error {
	var wg sync.WaitGroup

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	agent, err := NewAgent(cfg)
	if err != nil {
		log.Fatalf("create agent: %v", err)
	}

	fmt.Printf("Running agent on %s (protocol: %s)\n", cfg.ServerAddr, cfg.Protocol)

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal: %v. Shutting down gracefully...\n", sig)
		cancel()
	}()

	// HTTP workers
	if cfg.Protocol == "http" {
		for i := 0; i < cfg.RateLimit; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case m := <-agent.jobs:
						agent.httpReporter.send(m.MType, m.ID, m.String())
					}
				}
			}()
		}
	}

	// gRPC batch sender
	if cfg.Protocol == "grpc" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			gauges := make(map[string]float64)
			counters := make(map[string]int64)

			for {
				select {
				case <-ctx.Done():
					if len(gauges) > 0 || len(counters) > 0 {
						agent.grpcReporter.ReportBatch(gauges, counters)
					}
					return
				case <-ticker.C:
					if len(gauges) > 0 || len(counters) > 0 {
						agent.grpcReporter.ReportBatch(gauges, counters)
						gauges = make(map[string]float64)
						counters = make(map[string]int64)
					}
				case m := <-agent.jobs:
					if m.MType == storage.MetricTypeGauge && m.Value != nil {
						gauges[m.ID] = *m.Value
					} else if m.MType == storage.MetricTypeCounter && m.Delta != nil {
						counters[m.ID] = *m.Delta
					}
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
	
	if agent.grpcReporter != nil {
		agent.grpcReporter.Close()
	}

	fmt.Println("Agent stopped gracefully")
	return nil
}
