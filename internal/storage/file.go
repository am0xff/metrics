package storage

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Restore         bool
	FileStoragePath string
	StoreInterval   int
}

type DumpStorage struct {
	Gauges   map[string]Gauge   `json:"gauges,omitempty"`
	Counters map[string]Counter `json:"counters,omitempty"`
}

type FileStorage struct {
	ms  *MemoryStorage
	cfg Config
}

func NewFileStorage(cfg Config) (*FileStorage, error) {
	fs := &FileStorage{
		cfg: cfg,
		ms:  NewMemoryStorage(),
	}

	if !cfg.Restore {
		return fs, nil
	}

	data, err := os.ReadFile(cfg.FileStoragePath)
	if os.IsNotExist(err) {
		return fs, nil
	}
	if err != nil {
		return nil, err
	}

	var d DumpStorage
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}

	for k, v := range d.Gauges {
		fs.ms.SetGauge(k, v)
	}
	for k, v := range d.Counters {
		fs.ms.SetCounter(k, v)
	}

	return fs, nil
}

func (fs *FileStorage) GetGauge(key string) (Gauge, bool) {
	return fs.ms.GetGauge(key)
}

func (fs *FileStorage) KeysGauge() []string {
	return fs.ms.KeysGauge()
}

func (fs *FileStorage) GetCounter(key string) (Counter, bool) {
	return fs.ms.GetCounter(key)
}

func (fs *FileStorage) KeysCounter() []string {
	return fs.ms.KeysCounter()
}

func (fs *FileStorage) SetGauge(key string, value Gauge) {
	fs.ms.SetGauge(key, value)
	if fs.cfg.StoreInterval == 0 {
		if err := fs.Save(); err != nil {
			log.Printf("SetGauge %v", err)
		}
	}
}

func (fs *FileStorage) SetCounter(key string, value Counter) {
	fs.ms.SetCounter(key, value)
	if fs.cfg.StoreInterval == 0 {
		if err := fs.Save(); err != nil {
			log.Printf("SetCounter %v", err)
		}
	}
}

func (fs *FileStorage) MarshalJSON() ([]byte, error) {
	gauges := make(map[string]Gauge)
	for _, k := range fs.ms.KeysGauge() {
		if v, ok := fs.ms.GetGauge(k); ok {
			gauges[k] = v
		}
	}
	counters := make(map[string]Counter)
	for _, k := range fs.ms.KeysCounter() {
		if v, ok := fs.ms.GetCounter(k); ok {
			counters[k] = v
		}
	}

	return json.Marshal(DumpStorage{gauges, counters})
}

func (fs *FileStorage) Save() error {
	data, err := json.Marshal(fs)
	if err != nil {
		return err
	}

	return os.WriteFile(fs.cfg.FileStoragePath, data, 0666)
}
