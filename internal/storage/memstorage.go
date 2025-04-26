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

type DumpStorage struct {
	Gauges   map[string]Gauge   `json:"gauges,omitempty"`
	Counters map[string]Counter `json:"counters,omitempty"`
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

func (ms *MemStorage) MarshalJSON() ([]byte, error) {
	gauges := make(map[string]Gauge)
	for _, k := range ms.Gauges.Keys() {
		if v, ok := ms.Gauges.Get(k); ok {
			gauges[k] = v
		}
	}
	counters := make(map[string]Counter)
	for _, k := range ms.Counters.Keys() {
		if v, ok := ms.Counters.Get(k); ok {
			counters[k] = v
		}
	}

	return json.Marshal(DumpStorage{gauges, counters})
}

func (ms *MemStorage) Save(filename string) error {
	data, err := json.Marshal(ms)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0666)
}

func (ms *MemStorage) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var d DumpStorage
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}

	for k, v := range d.Gauges {
		ms.Gauges.Set(k, v)
	}
	for k, v := range d.Counters {
		ms.Counters.Add(k, v)
	}

	return nil
}
