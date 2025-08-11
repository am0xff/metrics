package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/am0xff/metrics/internal/logger"
	"github.com/am0xff/metrics/internal/middleware"
	"github.com/am0xff/metrics/internal/router"
	"github.com/am0xff/metrics/internal/storage"
	fstorage "github.com/am0xff/metrics/internal/storage/file"
	memstorage "github.com/am0xff/metrics/internal/storage/memory"
	pgstorage "github.com/am0xff/metrics/internal/storage/pg"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Run() error {
	ctx := context.Background()

	// Read config
	cfg, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if cfg.PprofEnabled {
		go func() {
			if err := http.ListenAndServe(cfg.PprofAddr, nil); err != nil {
				fmt.Printf("start pprof server: %v", err)
			}
		}()
	}

	// Connect to DB
	db, _ := sql.Open("pgx", cfg.DatabaseDSN)
	defer db.Close()

	// Init logger
	if err := logger.Initialize(); err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}

	var s storage.StorageProvider

	ms := memstorage.NewStorage()
	fs, _ := fstorage.NewStorage(ctx, fstorage.Config{
		FileStoragePath: cfg.FileStoragePath,
		Restore:         cfg.Restore,
		StoreInterval:   cfg.StoreInterval,
	})

	if cfg.DatabaseDSN != "" {
		var ds *pgstorage.PGStorage
		if cfg.DatabaseDSN != "" {
			ds = pgstorage.NewStorage(db)
			// Точка входа для создания таблиц
			if err := ds.Bootstrap(context.Background()); err != nil {
				return fmt.Errorf("bootstrap db storage: %w", err)
			}
			s = ds
		}
	} else if cfg.FileStoragePath != "" {
		s = fs
	} else {
		s = ms
	}

	r := router.SetupRoutes(s)

	handler := middleware.HashMiddleware(r, cfg.Key)
	handler = middleware.GzipMiddleware(handler, cfg.Key)
	handler = middleware.RSAMiddleware(handler, cfg.CryptoKey)
	handler = middleware.LoggerMiddleware(handler)

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
