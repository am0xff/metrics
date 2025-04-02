package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(s *Server) http.Handler {
	r := chi.NewRouter()
	r.Get("/", s.ListMetricsHandler)
	r.Get("/value/{type}/{name}", s.GetMetricHandler)
	r.Post("/update/{type}/{name}/{value}", s.UpdateMetricHandler)
	return r
}
