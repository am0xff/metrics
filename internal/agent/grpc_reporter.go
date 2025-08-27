package agent

import (
	"github.com/am0xff/metrics/internal/grpc/client"
)

type GRPCReporterWrapper struct {
	grpcClient *client.GRPCReporter
}

func NewGRPCReporter(cfg *ReporterConfig) (*GRPCReporterWrapper, error) {
	grpcClient, err := client.NewGRPCReporter(cfg.ServerAddr)
	if err != nil {
		return nil, err
	}

	return &GRPCReporterWrapper{
		grpcClient: grpcClient,
	}, nil
}

func (r *GRPCReporterWrapper) Report(gauges map[string]float64, counters map[string]int64) {
	r.grpcClient.Report(gauges, counters)
}

func (r *GRPCReporterWrapper) ReportBatch(gauges map[string]float64, counters map[string]int64) {
	r.grpcClient.ReportBatch(gauges, counters)
}

func (r *GRPCReporterWrapper) Close() error {
	return r.grpcClient.Close()
}
