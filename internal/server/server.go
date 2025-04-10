package server

import (
	"flag"
	"fmt"
	"github.com/am0xff/metrics/internal/storage"
	"github.com/caarlos0/env/v6"
	"log"
	"net/http"
)

type Server struct {
	Storage *storage.MemStorage
}

func NewServer(storage *storage.MemStorage) *Server {
	return &Server{Storage: storage}
}

func Run() {
	var config Config

	if err := env.Parse(&config); err != nil {
		log.Fatalf("Parse env: %v", err)
	}

	serverAddr := flag.String("a", config.ServerAddr, "HTTP сервер адрес")

	config = Config{
		ServerAddr: *serverAddr,
	}

	store := storage.NewMemStorage()
	srv := NewServer(store)
	router := SetupRoutes(srv)

	fmt.Println("Running server on", config.ServerAddr)
	if err := http.ListenAndServe(config.ServerAddr, router); err != nil {
		log.Fatal(err)
	}
}
