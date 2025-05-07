package memory

import (
	storage "github.com/am0xff/metrics/internal/storage"
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

func (m *MemStorage) GetGauge(key string) (storage.Gauge, bool) {
	return m.Gauges.Get(key)
}

func (m *MemStorage) GetCounter(key string) (storage.Counter, bool) {
	return m.Counters.Get(key)
}

func (m *MemStorage) SetGauge(key string, value storage.Gauge) {
	m.Gauges.Set(key, value)
}

func (m *MemStorage) SetCounter(key string, value storage.Counter) {
	m.Counters.Count(key, value)
}

func (m *MemStorage) KeysGauge() []string {
	return m.Gauges.Keys()
}

func (m *MemStorage) KeysCounter() []string {
	return m.Counters.Keys()
}
