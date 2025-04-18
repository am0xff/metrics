package server

import (
	"flag"
	"fmt"
	"github.com/am0xff/metrics/internal/logger"
	"github.com/am0xff/metrics/internal/storage"
	"github.com/caarlos0/env/v6"
	"net/http"
)

type Server struct {
	Storage *storage.MemStorage
}

func NewServer(storage *storage.MemStorage) *Server {
	return &Server{Storage: storage}
}

func Run() error {
	var config Config

	if err := env.Parse(&config); err != nil {
		return err
	}

	if err := logger.Initialize(); err != nil {
		return err
	}

	serverAddr := flag.String("a", config.ServerAddr, "HTTP сервер адрес")
	flag.Parse()

	config = Config{
		ServerAddr: *serverAddr,
	}

	store := storage.NewMemStorage()
	srv := NewServer(store)
	router := SetupRoutes(srv)

	fmt.Println("Running server on", config.ServerAddr)
	return http.ListenAndServe(config.ServerAddr, logger.WithLogger(router))
}
