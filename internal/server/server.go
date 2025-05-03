package server

import (
	"database/sql"
	"fmt"
	"github.com/am0xff/metrics/internal/handlers"
	"github.com/am0xff/metrics/internal/logger"
	"github.com/am0xff/metrics/internal/middleware"
	"github.com/am0xff/metrics/internal/router"
	"github.com/am0xff/metrics/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
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

	// Connect to DB
	db, _ := sql.Open("pgx", cfg.DatabaseDSN)
	defer db.Close()

	// Init logger
	if err := logger.Initialize(); err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}

	var s handlers.StorageProvider
	ms := storage.NewMemoryStorage()

	ds, err := storage.NewDBStorage(db)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}

	fs, err := storage.NewFileStorage(storage.Config{
		FileStoragePath: cfg.FileStoragePath,
		Restore:         cfg.Restore,
		StoreInterval:   cfg.StoreInterval,
	})
	if err != nil {
		return fmt.Errorf("load storage: %w", err)
	}

	if cfg.DatabaseDSN != "" {
		s = ds
	} else if cfg.FileStoragePath != "" {
		s = fs
	} else {
		s = ms
	}

	r := router.SetupRoutes(s, db)

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
