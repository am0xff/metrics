package server

import (
	"context"
	_ "strconv"

	prt "github.com/am0xff/metrics/api/proto"
	"github.com/am0xff/metrics/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricsServer struct {
	prt.UnimplementedMetricsServiceServer
	storage storage.StorageProvider
}

func NewMetricsServer(s storage.StorageProvider) *MetricsServer {
	return &MetricsServer{storage: s}
}

func (s *MetricsServer) UpdateMetric(ctx context.Context, req *prt.UpdateMetricRequest) (*prt.UpdateMetricResponse, error) {
	metric := req.GetMetric()

	switch metric.Type {
	case "gauge":
		if metric.Value == nil {
			return nil, status.Error(codes.InvalidArgument, "value required for gauge")
		}
		s.storage.SetGauge(ctx, metric.Id, storage.Gauge(*metric.Value))

		return &prt.UpdateMetricResponse{
			Metric: &prt.Metric{
				Id:    metric.Id,
				Type:  metric.Type,
				Value: metric.Value,
			},
		}, nil

	case "counter":
		if metric.Delta == nil {
			return nil, status.Error(codes.InvalidArgument, "delta required for counter")
		}
		s.storage.SetCounter(ctx, metric.Id, storage.Counter(*metric.Delta))

		return &prt.UpdateMetricResponse{
			Metric: &prt.Metric{
				Id:    metric.Id,
				Type:  metric.Type,
				Delta: metric.Delta,
			},
		}, nil

	default:
		return nil, status.Error(codes.InvalidArgument, "unknown metric type")
	}
}

func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *prt.UpdateMetricsRequest) (*prt.UpdateMetricsResponse, error) {
	for _, metric := range req.GetMetrics() {
		switch metric.Type {
		case "gauge":
			if metric.Value == nil {
				return nil, status.Error(codes.InvalidArgument, "value required for gauge")
			}
			s.storage.SetGauge(ctx, metric.Id, storage.Gauge(*metric.Value))

		case "counter":
			if metric.Delta == nil {
				return nil, status.Error(codes.InvalidArgument, "delta required for counter")
			}
			s.storage.SetCounter(ctx, metric.Id, storage.Counter(*metric.Delta))

		default:
			return nil, status.Error(codes.InvalidArgument, "unknown metric type")
		}
	}

	return &prt.UpdateMetricsResponse{Success: true}, nil
}

func (s *MetricsServer) GetMetric(ctx context.Context, req *prt.GetMetricRequest) (*prt.GetMetricResponse, error) {
	switch req.Type {
	case "gauge":
		value, exists := s.storage.GetGauge(ctx, req.Id)
		if !exists {
			return nil, status.Error(codes.NotFound, "metric not found")
		}

		v := float64(value)
		return &prt.GetMetricResponse{
			Metric: &prt.Metric{
				Id:    req.Id,
				Type:  req.Type,
				Value: &v,
			},
		}, nil

	case "counter":
		value, exists := s.storage.GetCounter(ctx, req.Id)
		if !exists {
			return nil, status.Error(codes.NotFound, "metric not found")
		}

		d := int64(value)
		return &prt.GetMetricResponse{
			Metric: &prt.Metric{
				Id:    req.Id,
				Type:  req.Type,
				Delta: &d,
			},
		}, nil

	default:
		return nil, status.Error(codes.InvalidArgument, "unknown metric type")
	}
}

func (s *MetricsServer) GetMetrics(ctx context.Context, req *prt.GetMetricsRequest) (*prt.GetMetricsResponse, error) {
	var metrics []*prt.Metric

	for _, key := range s.storage.KeysGauge(ctx) {
		if value, exists := s.storage.GetGauge(ctx, key); exists {
			v := float64(value)
			metrics = append(metrics, &prt.Metric{
				Id:    key,
				Type:  "gauge",
				Value: &v,
			})
		}
	}

	for _, key := range s.storage.KeysCounter(ctx) {
		if value, exists := s.storage.GetCounter(ctx, key); exists {
			d := int64(value)
			metrics = append(metrics, &prt.Metric{
				Id:    key,
				Type:  "counter",
				Delta: &d,
			})
		}
	}

	return &prt.GetMetricsResponse{Metrics: metrics}, nil
}

func (s *MetricsServer) Ping(ctx context.Context, req *prt.PingRequest) (*prt.PingResponse, error) {
	if err := s.storage.Ping(ctx); err != nil {
		return &prt.PingResponse{Success: false}, nil
	}
	return &prt.PingResponse{Success: true}, nil
}
