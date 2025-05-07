package storage

import (
	"database/sql"
	storage "github.com/am0xff/metrics/internal/storage"
	memstorage "github.com/am0xff/metrics/internal/storage/memory"
	"log"
)

type DBStorage struct {
	ms *memstorage.MemStorage
	db *sql.DB
}

func NewStorage(db *sql.DB) (*DBStorage, error) {
	// Создаем таблицы, если их нет
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS gauges (
			key TEXT PRIMARY KEY,
			value DOUBLE PRECISION NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS counters (
			key TEXT PRIMARY KEY,
			value BIGINT NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	return &DBStorage{
		ms: memstorage.NewStorage(),
		db: db,
	}, nil
}

func (d *DBStorage) SetGauge(key string, value storage.Gauge) {
	d.ms.SetGauge(key, value)

	_, err := d.db.Exec(`
		INSERT INTO gauges (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
	`, key, float64(value))

	if err != nil {
		log.Printf("DBStorage.SetGauge exec error: %v", err)
	}
}

func (d *DBStorage) GetGauge(key string) (storage.Gauge, bool) {
	var v float64
	err := d.db.QueryRow(`
		SELECT value FROM gauges WHERE key = $1
	`, key).Scan(&v)

	if err != nil {
		log.Printf("DBStorage.GetGauge query error: %v", err)
		// fallback to memory
		return d.ms.GetGauge(key)
	}
	return storage.Gauge(v), true
}

func (d *DBStorage) KeysGauge() []string {
	rows, err := d.db.Query(`SELECT key FROM gauges`)
	if err != nil {
		log.Printf("DBStorage.KeysGauge query error: %v", err)
		return d.ms.KeysGauge()
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
		return d.ms.KeysGauge()
	}

	return keys
}

func (d *DBStorage) SetCounter(key string, value storage.Counter) {
	d.ms.SetCounter(key, value)

	_, err := d.db.Exec(`
		INSERT INTO counters (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = counters.value + EXCLUDED.value
	`, key, int64(value))

	if err != nil {
		log.Printf("DBStorage.SetCounter exec error: %v", err)
	}
}

func (d *DBStorage) GetCounter(key string) (storage.Counter, bool) {
	var v int64
	err := d.db.QueryRow(`
		SELECT value FROM counters WHERE key = $1
	`, key).Scan(&v)

	if err != nil {
		log.Printf("DBStorage.GetCounter query error: %v", err)
		return d.ms.GetCounter(key)
	}
	return storage.Counter(v), true
}

func (d *DBStorage) KeysCounter() []string {
	rows, err := d.db.Query(`SELECT key FROM counters`)
	if err != nil {
		log.Printf("DBStorage.KeysCounter query error: %v", err)
		return d.ms.KeysCounter()
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
		return d.ms.KeysCounter()
	}

	return keys
}
