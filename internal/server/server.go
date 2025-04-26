package server

import (
	"fmt"
	"github.com/am0xff/metrics/internal/logger"
	"github.com/am0xff/metrics/internal/middleware"
	"github.com/am0xff/metrics/internal/router"
	"github.com/am0xff/metrics/internal/storage"
	"log"
	"net/http"
	"time"
)

func Run() error {
	// Read config
	cfg, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Init logger
	if err := logger.Initialize(); err != nil {
		return err
	}

	s := storage.NewMemStorage()

	if cfg.Restore {
		if err := s.Load(cfg.FileStoragePath); err != nil {
			return fmt.Errorf("load storage: %w", err)
		}
	}

	r := router.SetupRoutes(s)

	handler := logger.WithLogger(r)
	handler = middleware.GzipMiddleware(handler)

	go func() {
		for {
			if err := s.Save(cfg.FileStoragePath); err != nil {
				log.Print(err)
			}
			time.Sleep(time.Duration(cfg.StoreInterval) * time.Second)
		}
	}()

	fmt.Println("Running server on", cfg.ServerAddr)
	return http.ListenAndServe(cfg.ServerAddr, handler)
}
