package storage

import (
	"encoding/json"
	"os"
)

type Gauge float64
type Counter int64

type GaugeStorage struct {
	storage Storage[Gauge]
}

type CounterStorage struct {
	storage Storage[Counter]
}

type MemStorage struct {
	Gauges   *Storage[Gauge]
	Counters *Storage[Counter]
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   NewStorage[Gauge](),
		Counters: NewStorage[Counter](),
	}
}

func (gs *GaugeStorage) Set(key string, value Gauge) {
	gs.storage.Set(key, value)
}

func (gs *GaugeStorage) Get(key string) (Gauge, bool) {
	return gs.storage.Get(key)
}

func (gs *GaugeStorage) Keys() []string {
	return gs.storage.Keys()
}

func (cs *CounterStorage) Add(key string, value Counter) {
	cs.storage.Add(key, value)
}

func (cs *CounterStorage) Get(key string) (Counter, bool) {
	return cs.storage.Get(key)
}

func (cs *CounterStorage) Keys() []string {
	return cs.storage.Keys()
}

func (ms *MemStorage) Save(filename string) error {
	// Collect data
	gauges := make(map[string]Gauge, 0)
	for _, key := range ms.Gauges.Keys() {
		if v, ok := ms.Gauges.Get(key); ok {
			gauges[key] = v
		}
	}
	counters := make(map[string]Counter, 0)
	for _, key := range ms.Counters.Keys() {
		if v, ok := ms.Counters.Get(key); ok {
			counters[key] = v
		}
	}

	dump := struct {
		Gauges   map[string]Gauge
		Counters map[string]Counter
	}{
		Gauges:   gauges,
		Counters: counters,
	}

	data, err := json.Marshal(&dump)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0666)
}

func (ms *MemStorage) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, ms); err != nil {
		return err
	}

	return nil
}
