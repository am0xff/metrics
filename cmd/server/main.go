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

	mux.HandleFunc("/update/gauge/", handler.APIGauge)
	mux.HandleFunc("/update/counter/", handler.APICounter)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
