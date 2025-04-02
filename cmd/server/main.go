package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/am0xff/metrics/internal/server"
	"github.com/am0xff/metrics/internal/storage"
)

func main() {
	addr := flag.String("a", "localhost:8080", "Адрес HTTP сервера")
	flag.Parse()

	store := storage.NewMemStorage()
	srv := server.NewServer(store)
	router := server.SetupRoutes(srv)

	log.Printf("Запуск сервера на %s\n", *addr)
	if err := http.ListenAndServe(*addr, router); err != nil {
		log.Fatal(err)
	}
}
