package router

import (
	"github.com/am0xff/metrics/internal/handlers"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func SetupRoutes(sp handlers.StorageProvider) http.Handler {
	r := chi.NewRouter()

	handler := handlers.NewHandler(sp)

	r.Get("/", handler.GetMetrics)
	r.Post("/value/", handler.POSTGetMetric)
	r.Post("/update/", handler.POSTUpdateMetric)
	r.Get("/value/{type}/{name}", handler.GETGetMetric)
	r.Post("/update/{type}/{name}/{value}", handler.GETUpdateMetric)
	return r
}
