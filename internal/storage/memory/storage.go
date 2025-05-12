package memory

import (
	"context"
	"github.com/am0xff/metrics/internal/storage"
)

type MemStorage struct {
	Gauges   *storage.Storage[storage.Gauge]
	Counters *storage.Storage[storage.Counter]
}

func NewStorage() *MemStorage {
	return &MemStorage{
		Gauges:   storage.NewStorage[storage.Gauge](),
		Counters: storage.NewStorage[storage.Counter](),
	}
}

func (m *MemStorage) GetGauge(_ context.Context, key string) (storage.Gauge, bool) {
	return m.Gauges.Get(key)
}

func (m *MemStorage) GetCounter(_ context.Context, key string) (storage.Counter, bool) {
	return m.Counters.Get(key)
}

func (m *MemStorage) SetGauge(_ context.Context, key string, value storage.Gauge) {
	m.Gauges.Set(key, value)
}

func (m *MemStorage) SetCounter(_ context.Context, key string, value storage.Counter) {
	m.Counters.Count(key, value)
}

func (m *MemStorage) KeysGauge(_ context.Context) []string {
	return m.Gauges.Keys()
}

func (m *MemStorage) KeysCounter(_ context.Context) []string {
	return m.Counters.Keys()
}

func (m *MemStorage) Ping(_ context.Context) error {
	return nil
}
