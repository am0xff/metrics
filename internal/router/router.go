package router

import (
	"database/sql"
	"github.com/am0xff/metrics/internal/handlers"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func SetupRoutes(sp handlers.StorageProvider, db *sql.DB) http.Handler {
	r := chi.NewRouter()

	handler := handlers.NewHandler(sp, db)

	r.Get("/", handler.GetMetrics)
	r.Get("/ping", handler.Ping)
	r.Post("/value/", handler.POSTGetMetric)
	r.Post("/update/", handler.POSTUpdateMetric)
	r.Post("/updates/", handler.POSTUpdatesMetrics)
	r.Get("/value/{type}/{name}", handler.GETGetMetric)
	r.Post("/update/{type}/{name}/{value}", handler.GETUpdateMetric)
	return r
}
