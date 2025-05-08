package file

import (
	"context"
	"encoding/json"
	"github.com/am0xff/metrics/internal/storage"
	memstorage "github.com/am0xff/metrics/internal/storage/memory"
	"github.com/am0xff/metrics/internal/utils"
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
	ctx context.Context
}

func NewStorage(ctx context.Context, cfg Config) (*FileStorage, error) {
	fs := &FileStorage{
		cfg: cfg,
		ms:  memstorage.NewStorage(),
		ctx: ctx,
	}

	if !cfg.Restore {
		return fs, nil
	}

	//data, err := os.ReadFile(cfg.FileStoragePath)
	//if os.IsNotExist(err) {
	//	return fs, nil
	//}
	//if err != nil {
	//	return nil, err
	//}
	var data []byte
	if err := utils.Call(ctx, func() error {
		var err error
		data, err = os.ReadFile(cfg.FileStoragePath)
		return err
	}); err != nil {
		if os.IsNotExist(err) {
			return fs, nil
		}
		return nil, err
	}

	var d DumpStorage
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}

	for k, v := range d.Gauges {
		fs.ms.SetGauge(ctx, k, v)
	}
	for k, v := range d.Counters {
		fs.ms.SetCounter(ctx, k, v)
	}

	return fs, nil
}

func (fs *FileStorage) GetGauge(ctx context.Context, key string) (storage.Gauge, bool) {
	return fs.ms.GetGauge(ctx, key)
}

func (fs *FileStorage) KeysGauge(ctx context.Context) []string {
	return fs.ms.KeysGauge(ctx)
}

func (fs *FileStorage) GetCounter(ctx context.Context, key string) (storage.Counter, bool) {
	return fs.ms.GetCounter(ctx, key)
}

func (fs *FileStorage) KeysCounter(ctx context.Context) []string {
	return fs.ms.KeysCounter(ctx)
}

func (fs *FileStorage) SetGauge(ctx context.Context, key string, value storage.Gauge) {
	fs.ms.SetGauge(ctx, key, value)
	if fs.cfg.StoreInterval == 0 {
		if err := fs.Save(); err != nil {
			log.Printf("SetGauge %v", err)
		}
	}
}

func (fs *FileStorage) SetCounter(ctx context.Context, key string, value storage.Counter) {
	fs.ms.SetCounter(ctx, key, value)
	if fs.cfg.StoreInterval == 0 {
		if err := fs.Save(); err != nil {
			log.Printf("SetCounter %v", err)
		}
	}
}

func (fs *FileStorage) MarshalJSON() ([]byte, error) {
	gauges := make(map[string]storage.Gauge)
	for _, k := range fs.ms.KeysGauge(fs.ctx) {
		if v, ok := fs.ms.GetGauge(fs.ctx, k); ok {
			gauges[k] = v
		}
	}
	counters := make(map[string]storage.Counter)
	for _, k := range fs.ms.KeysCounter(fs.ctx) {
		if v, ok := fs.ms.GetCounter(fs.ctx, k); ok {
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

	return utils.Call(context.Background(), func() error {
		return os.WriteFile(fs.cfg.FileStoragePath, data, 0666)
	})
}
