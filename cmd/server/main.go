package main

import (
	"github.com/am0xff/metrics/internal/handlers"
	"github.com/am0xff/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	store := storage.NewMemStorage()
	handler := handlers.NewHandler(store)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.RootHandle)
		r.Route("/value", func(r chi.Router) {
			r.Route("/gauge", func(r chi.Router) {
				r.Get("/{metric_name}", handler.GetGaugeMetric)
			})
			r.Route("/counter", func(r chi.Router) {
				r.Get("/{metric_name}", handler.GetCounterMetric)
			})
		})
		r.Route("/update", func(r chi.Router) {
			r.Route("/gauge", func(r chi.Router) {
				r.Route("/{metric_name}", func(r chi.Router) {
					r.Post("/{metric_value}", handler.UpdateGaugeMetric)
				})
			})
			r.Route("/counter", func(r chi.Router) {
				r.Route("/{metric_name}", func(r chi.Router) {
					r.Post("/{metric_value}", handler.UpdateCounterMetric)
				})
			})
		})
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
