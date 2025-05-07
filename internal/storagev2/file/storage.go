package file

import (
	"encoding/json"
	storage "github.com/am0xff/metrics/internal/storagev2"
	memstorage "github.com/am0xff/metrics/internal/storagev2/memory"
	"log"
	"os"
)

type Config struct {
	Restore         bool
	FileStoragePath string
	StoreInterval   int
}

type DumpStorage struct {
	Gauges   map[string]storage.Gauge   `json:"gauges,omitempty"`
	Counters map[string]storage.Counter `json:"counters,omitempty"`
}

type FileStorage struct {
	ms  *memstorage.MemStorage
	cfg Config
}

func NewStorage(cfg Config) (*FileStorage, error) {
	fs := &FileStorage{
		cfg: cfg,
		ms:  memstorage.NewStorage(),
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

func (fs *FileStorage) GetGauge(key string) (storage.Gauge, bool) {
	return fs.ms.GetGauge(key)
}

func (fs *FileStorage) KeysGauge() []string {
	return fs.ms.KeysGauge()
}

func (fs *FileStorage) GetCounter(key string) (storage.Counter, bool) {
	return fs.ms.GetCounter(key)
}

func (fs *FileStorage) KeysCounter() []string {
	return fs.ms.KeysCounter()
}

func (fs *FileStorage) SetGauge(key string, value storage.Gauge) {
	fs.ms.SetGauge(key, value)
	if fs.cfg.StoreInterval == 0 {
		if err := fs.Save(); err != nil {
			log.Printf("SetGauge %v", err)
		}
	}
}

func (fs *FileStorage) SetCounter(key string, value storage.Counter) {
	fs.ms.SetCounter(key, value)
	if fs.cfg.StoreInterval == 0 {
		if err := fs.Save(); err != nil {
			log.Printf("SetCounter %v", err)
		}
	}
}

func (fs *FileStorage) MarshalJSON() ([]byte, error) {
	gauges := make(map[string]storage.Gauge)
	for _, k := range fs.ms.KeysGauge() {
		if v, ok := fs.ms.GetGauge(k); ok {
			gauges[k] = v
		}
	}
	counters := make(map[string]storage.Counter)
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
