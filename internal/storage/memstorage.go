package storage

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

func (cs *CounterStorage) Set(key string, value Counter) {
	cs.storage.Set(key, value)
}

func (cs *CounterStorage) Get(key string) (Counter, bool) {
	return cs.storage.Get(key)
}

func (cs *CounterStorage) Keys() []string {
	return cs.storage.Keys()
}
