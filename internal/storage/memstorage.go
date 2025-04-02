package storage

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MemStorage) UpdateGauge(name string, value float64) error {
	m.gauges[name] = value
	return nil
}

func (m *MemStorage) UpdateCounter(name string, value int64) error {
	m.counters[name] += value
	return nil
}

func (m *MemStorage) GetGauge(name string) (float64, bool) {
	v, ok := m.gauges[name]
	return v, ok
}

func (m *MemStorage) GetCounter(name string) (int64, bool) {
	v, ok := m.counters[name]
	return v, ok
}

func (m *MemStorage) GetAllGauges() map[string]float64 {
	copyMap := make(map[string]float64)
	for k, v := range m.gauges {
		copyMap[k] = v
	}
	return copyMap
}

func (m *MemStorage) GetAllCounters() map[string]int64 {
	copyMap := make(map[string]int64)
	for k, v := range m.counters {
		copyMap[k] = v
	}
	return copyMap
}
