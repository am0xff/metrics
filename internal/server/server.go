package server

import "github.com/am0xff/metrics/internal/metrics"

type Server struct {
	Storage metrics.MetricsStorage
}

func NewServer(storage metrics.MetricsStorage) *Server {
	return &Server{Storage: storage}
}
