package server

import (
	"fmt"
	"github.com/am0xff/metrics/internal/logger"
	"github.com/am0xff/metrics/internal/middleware"
	"github.com/am0xff/metrics/internal/router"
	"github.com/am0xff/metrics/internal/storage"
	"log"
	"net/http"
)

func Run() error {
	// Read config
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Init logger
	if err := logger.Initialize(); err != nil {
		return err
	}

	s := storage.NewMemStorage()
	r := router.SetupRoutes(s)

	handler := logger.WithLogger(r)
	handler = middleware.GzipMiddleware(handler)

	fmt.Println("Running server on", cfg.ServerAddr)
	return http.ListenAndServe(cfg.ServerAddr, handler)
}
