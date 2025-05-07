package storage

type Gauge float64
type Counter int64

type MetricType string

const (
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeCounter MetricType = "counter"
)

type StorageProvider interface {
	GetGauge(key string) (Gauge, bool)
	GetCounter(key string) (Counter, bool)
	KeysGauge() []string
	SetGauge(key string, value Gauge)
	SetCounter(key string, value Counter)
	KeysCounter() []string
}

type Storage[T interface{ Gauge | Counter }] struct {
	data map[string]T
}

func NewStorage[T interface{ Gauge | Counter }]() *Storage[T] {
	return &Storage[T]{
		data: make(map[string]T),
	}
}

func (s *Storage[T]) Get(key string) (T, bool) {
	val, ok := s.data[key]
	return val, ok
}

func (s *Storage[T]) Set(key string, val T) {
	s.data[key] = val
}

func (s *Storage[T]) Keys() []string {
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

func (s *Storage[T]) Count(key string, value T) {
	s.data[key] += value
}
