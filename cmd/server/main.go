package main

import (
	"github.com/am0xff/metrics/internal/handlers"
	"github.com/am0xff/metrics/internal/storage"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	store := storage.NewMemStorage()
	handler := handlers.NewHandler(store)

	mux.HandleFunc("/update/gauge/", handler.ApiGauge)
	mux.HandleFunc("/update/counter/", handler.ApiCounter)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
