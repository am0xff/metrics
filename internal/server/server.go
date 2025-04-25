package server

import (
	"flag"
	"fmt"
	"github.com/am0xff/metrics/internal/logger"
	"github.com/am0xff/metrics/internal/middleware"
	"github.com/am0xff/metrics/internal/router"
	"github.com/am0xff/metrics/internal/storage"
	"github.com/caarlos0/env/v6"
	"net/http"
)

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

	s := storage.NewMemStorage()
	r := router.SetupRoutes(s)

	handler := logger.WithLogger(r)
	handler = middleware.GzipMiddleware(handler)

	fmt.Println("Running server on", config.ServerAddr)
	return http.ListenAndServe(config.ServerAddr, handler)
}
