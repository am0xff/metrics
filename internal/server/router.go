package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(s *Server) http.Handler {
	r := chi.NewRouter()
	r.Get("/", s.ListMetricsHandler)
	// old
	//r.Get("/value/{type}/{name}", s.GetMetricHandler)
	//r.Post("/update/{type}/{name}/{value}", s.UpdateMetricHandler)
	// New
	r.Get("/value", s.GetMetricHandler)
	r.Post("/update", s.UpdateMetricHandler)
	return r
}
