package storage

import (
	"context"
	"database/sql"
	"github.com/am0xff/metrics/internal/storage"
	memstorage "github.com/am0xff/metrics/internal/storage/memory"
	"log"
)

type PGStorage struct {
	ms *memstorage.MemStorage
	db *sql.DB
}

func NewStorage(db *sql.DB) *PGStorage {
	return &PGStorage{
		db: db,
		ms: memstorage.NewStorage(),
	}
}

func (pgs *PGStorage) Bootstrap(ctx context.Context) error {
	tx, err := pgs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := pgs.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS gauges (
			key TEXT PRIMARY KEY,
			value DOUBLE PRECISION NOT NULL
		)
	`); err != nil {
		return err
	}

	if _, err = pgs.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS counters (
			key TEXT PRIMARY KEY,
			value BIGINT NOT NULL
		)
	`); err != nil {
		return err
	}

	return tx.Commit()
}

func (pgs *PGStorage) SetGauge(ctx context.Context, key string, value storage.Gauge) {
	pgs.ms.SetGauge(ctx, key, value)

	_, err := pgs.db.ExecContext(ctx, `
		INSERT INTO gauges (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
	`, key, float64(value))

	if err != nil {
		log.Printf("DBStorage.SetGauge exec error: %v", err)
	}
}

func (pgs *PGStorage) GetGauge(ctx context.Context, key string) (storage.Gauge, bool) {
	var v float64
	err := pgs.db.QueryRowContext(ctx, `
		SELECT value FROM gauges WHERE key = $1
	`, key).Scan(&v)

	if err != nil {
		log.Printf("DBStorage.GetGauge query error: %v", err)
		// fallback to memory
		return d.ms.GetGauge(ctx, key)
	}
	return storage.Gauge(v), true
}

func (pgs *PGStorage) KeysGauge(ctx context.Context) []string {
	rows, err := pgs.db.QueryContext(ctx, `SELECT key FROM gauges`)
	if err != nil {
		log.Printf("DBStorage.KeysGauge query error: %v", err)
		return pgs.ms.KeysGauge(ctx)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			log.Printf("DBStorage.KeysGauge scan error: %v", err)
			continue
		}
		keys = append(keys, k)
	}

	if err := rows.Err(); err != nil {
		log.Printf("DBStorage.KeysGauge rows iteration error: %v", err)
		return d.ms.KeysGauge(ctx)
	}

	return keys
}

func (pgs *PGStorage) SetCounter(ctx context.Context, key string, value storage.Counter) {
	pgs.ms.SetCounter(ctx, key, value)

	_, err := pgs.db.ExecContext(ctx, `
		INSERT INTO counters (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = counters.value + EXCLUDED.value
	`, key, int64(value))

	if err != nil {
		log.Printf("DBStorage.SetCounter exec error: %v", err)
	}
}

func (pgs *PGStorage) GetCounter(ctx context.Context, key string) (storage.Counter, bool) {
	var v int64
	err := pgs.db.QueryRowContext(ctx, `
		SELECT value FROM counters WHERE key = $1
	`, key).Scan(&v)

	if err != nil {
		log.Printf("DBStorage.GetCounter query error: %v", err)
		return pgs.ms.GetCounter(ctx, key)
	}
	return storage.Counter(v), true
}

func (pgs *PGStorage) KeysCounter(ctx context.Context) []string {
	rows, err := pgs.db.QueryContext(ctx, `SELECT key FROM counters`)
	if err != nil {
		log.Printf("DBStorage.KeysCounter query error: %v", err)
		return pgs.ms.KeysCounter(ctx)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			log.Printf("DBStorage.KeysCounter scan error: %v", err)
			continue
		}
		keys = append(keys, k)
	}

	if err := rows.Err(); err != nil {
		log.Printf("DBStorage.KeysGauge rows iteration error: %v", err)
		return pgs.ms.KeysCounter(ctx)
	}

	return keys
}
