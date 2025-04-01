package main

import (
	"fmt"
	"github.com/am0xff/metrics/internal/handlers"
	"github.com/am0xff/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	// обрабатываем аргументы командной строки
	parseFlags()

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	fmt.Println("Running server on", flagRunAddr)

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
			r.Route("/{metric_type}", func(r chi.Router) {
				r.Route("/{metric_name}", func(r chi.Router) {
					r.Post("/{metric_value}", handler.UpdateMetric)
				})
			})
		})
	})

	return http.ListenAndServe(flagRunAddr, r)
}
