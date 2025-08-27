package client

import (
	"context"
	"log"
	"time"

	"github.com/am0xff/metrics/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCReporter struct {
	client gen.MetricsServiceClient
	conn   *grpc.ClientConn
}

func NewGRPCReporter(addr string) (*GRPCReporter, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := gen.NewMetricsServiceClient(conn)

	return &GRPCReporter{
		client: client,
		conn:   conn,
	}, nil
}

func (r *GRPCReporter) Close() error {
	return r.conn.Close()
}

func (r *GRPCReporter) Report(gauges map[string]float64, counters map[string]int64) {
	for name, value := range gauges {
		r.sendGauge(name, value)
	}
	for name, delta := range counters {
		r.sendCounter(name, delta)
	}
}

func (r *GRPCReporter) ReportBatch(gauges map[string]float64, counters map[string]int64) {
	var metrics []*gen.Metric

	for name, value := range gauges {
		v := value
		metrics = append(metrics, &gen.Metric{
			Id:    name,
			Type:  "gauge",
			Value: &v,
		})
	}

	for name, delta := range counters {
		d := delta
		metrics = append(metrics, &gen.Metric{
			Id:    name,
			Type:  "counter",
			Delta: &d,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.client.UpdateMetrics(ctx, &gen.UpdateMetricsRequest{
		Metrics: metrics,
	})
	if err != nil {
		log.Printf("gRPC batch update failed: %v", err)
	}
}

func (r *GRPCReporter) sendGauge(name string, value float64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.client.UpdateMetric(ctx, &gen.UpdateMetricRequest{
		Metric: &gen.Metric{
			Id:    name,
			Type:  "gauge",
			Value: &value,
		},
	})
	if err != nil {
		log.Printf("gRPC gauge update failed for %s: %v", name, err)
	}
}

func (r *GRPCReporter) sendCounter(name string, delta int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.client.UpdateMetric(ctx, &gen.UpdateMetricRequest{
		Metric: &gen.Metric{
			Id:    name,
			Type:  "counter",
			Delta: &delta,
		},
	})
	if err != nil {
		log.Printf("gRPC counter update failed for %s: %v", name, err)
	}
}
