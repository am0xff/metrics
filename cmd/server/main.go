package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/am0xff/metrics/internal/server"
	"github.com/am0xff/metrics/internal/storage"
)

var flagRunAddr string

func main() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}

	store := storage.NewMemStorage()
	srv := server.NewServer(store)
	router := server.SetupRoutes(srv)

	fmt.Println("Running server on", flagRunAddr)
	if err := http.ListenAndServe(flagRunAddr, router); err != nil {
		log.Fatal(err)
	}
}
