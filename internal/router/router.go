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
	r.Post("/update/", handler.POSTUpdateMetric)
	r.Post("/updates/", handler.POSTUpdatesMetrics)
	r.Get("/value/gauge/{name}", handler.GETGetMetricGauge)
	r.Get("/value/counter/{name}", handler.GETGetMetricCounter)
	r.Get("/update/gauge/{name}/{value}", handler.GETUpdateGauge)
	r.Get("/update/counter/{name}/{value}", handler.GETUpdateCounter)
	return r
}
