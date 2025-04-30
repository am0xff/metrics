package storage

type MetricType string

const (
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeCounter MetricType = "counter"
)

type Gauge float64
type Counter int64

type GaugeStorage struct {
	storage Storage[Gauge]
}

type CounterStorage struct {
	storage Storage[Counter]
}

type MemoryStorage struct {
	Gauges   *Storage[Gauge]
	Counters *Storage[Counter]
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		Gauges:   NewStorage[Gauge](),
		Counters: NewStorage[Counter](),
	}
}

func (ms *MemoryStorage) SetGauge(key string, value Gauge) {
	ms.Gauges.Set(key, value)
}

func (ms *MemoryStorage) GetGauge(key string) (Gauge, bool) {
	return ms.Gauges.Get(key)
}

func (ms *MemoryStorage) KeysGauge() []string {
	return ms.Gauges.Keys()
}

func (ms *MemoryStorage) SetCounter(key string, value Counter) {
	ms.Counters.Add(key, value)
}

func (ms *MemoryStorage) GetCounter(key string) (Counter, bool) {
	return ms.Counters.Get(key)
}

func (ms *MemoryStorage) KeysCounter() []string {
	return ms.Counters.Keys()
}
