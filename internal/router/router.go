package router

import (
	"github.com/am0xff/metrics/internal/handlers"
	"github.com/am0xff/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func SetupRoutes(s *storage.MemStorage) http.Handler {
	r := chi.NewRouter()

	handler := handlers.NewHandler(s)

	r.Get("/", handler.GetMetrics)
	r.Post("/value", handler.GetMetric)
	r.Post("/update", handler.UpdateMetric)
	return r
}
