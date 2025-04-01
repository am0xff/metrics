package storage

type gauge float64
type counter int64

type MemStorage struct {
	Gauge   map[string]gauge
	Counter map[string]counter
}

func NewMemStorage() *MemStorage {
	return &MemStorage{Gauge: make(map[string]gauge), Counter: make(map[string]counter)}
}

func (m *MemStorage) SetGauge(name string, value float64) {
	m.Gauge[name] = gauge(value)
}

func (m *MemStorage) SetCounter(name string, value int64) {
	m.Counter[name] += counter(value)
}

func (m *MemStorage) GetGauge(name string) (float64, bool) {
	val, ok := m.Gauge[name]
	return float64(val), ok
}

func (m *MemStorage) GetCounter(name string) (int64, bool) {
	val, ok := m.Counter[name]
	return int64(val), ok
}

func (m *MemStorage) GaugeValues() map[string]gauge {
	return m.Gauge
}

func (m *MemStorage) CounterValues() map[string]counter {
	return m.Counter
}
