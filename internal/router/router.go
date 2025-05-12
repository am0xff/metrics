package router

import (
	"github.com/am0xff/metrics/internal/handlers"
	"github.com/am0xff/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func SetupRoutes(sp storage.StorageProvider) http.Handler {
	r := chi.NewRouter()

	handler := handlers.NewHandler(sp)

	r.Get("/", handler.GetMetrics)
	r.Get("/ping", handler.Ping)
	r.Post("/value/", handler.POSTGetMetric)
	r.Post("/update/gauge/", handler.POSTUpdateMetricGauge)
	r.Post("/update/counter/", handler.POSTUpdateMetricCounter)
	r.Post("/updates/", handler.POSTUpdatesMetrics)
	r.Get("/value/{type}/{name}", handler.GETGetMetric)
	r.Get("/update/{type}/{name}/{value}", handler.GETUpdateMetric)
	return r
}
