package storage

type gauge float64
type counter int64

type MemStorage struct {
	storageGauge   map[string]gauge
	storageCounter map[string]counter
}

func NewMemStorage() *MemStorage {
	return &MemStorage{storageGauge: make(map[string]gauge), storageCounter: make(map[string]counter)}
}

func (m *MemStorage) SetGauge(name string, value float64) {
	m.storageGauge[name] = gauge(value)
}

func (m *MemStorage) SetCounter(name string, value int64) {
	m.storageCounter[name] += counter(value)
}
