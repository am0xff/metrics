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

	fs, err := storage.NewFileStorage(storage.Config{
		FileStoragePath: cfg.FileStoragePath,
		Restore:         cfg.Restore,
		StoreInterval:   cfg.StoreInterval,
	})

	if err != nil {
		return fmt.Errorf("load storage: %w", err)
	}

	r := router.SetupRoutes(fs)

	handler := middleware.LoggerMiddleware(r)
	handler = middleware.GzipMiddleware(handler)

	if cfg.StoreInterval != 0 {
		go func() {
			for {
				if err := fs.Save(); err != nil {
					log.Printf("Save storage to the file: %v", err)
				}
				time.Sleep(time.Duration(cfg.StoreInterval) * time.Second)
			}
		}()
	}

	fmt.Println("Running server on", cfg.ServerAddr)
	return http.ListenAndServe(cfg.ServerAddr, handler)
}
