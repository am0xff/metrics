// Package storage предоставляет интерфейсы и типы для хранения метрик.
// Пакет поддерживает работу с двумя типами метрик: gauge и counter.
// Предоставляет универсальный интерфейс StorageProvider для различных реализаций хранилища.
package storage

import (
	"context"
)

// Gauge представляет тип метрики для измерения текущего значения показателя.
// Используется для метрик типа gauge, таких как использование CPU, памяти,
// температура и другие измеряемые величины.
//
// Пример: использование CPU 85.5%, температура 23.7°C
type Gauge float64

// Counter представляет тип метрики для подсчета событий.
// Используется для метрик типа counter, таких как количество запросов,
// ошибок, обработанных сообщений и других счетчиков.
//
// Пример: количество HTTP запросов 1234, количество ошибок 5
type Counter int64

// MetricType определяет тип метрики.
type MetricType string

const (
	// MetricTypeGauge представляет тип метрики gauge.
	// Используется для измеряемых показателей, которые могут увеличиваться и уменьшаться.
	MetricTypeGauge MetricType = "gauge"

	// MetricTypeCounter представляет тип метрики counter.
	// Используется для счетчиков, которые только увеличиваются.
	MetricTypeCounter MetricType = "counter"
)

// StorageProvider определяет интерфейс для работы с хранилищем метрик.
// Интерфейс поддерживает операции получения, установки значений метрик
// и получения списка ключей для обоих типов метрик.
//
// Пример реализации может использовать память, файловую систему,
// базу данных или другие системы хранения.
//
// Пример использования:
//
//	var storage StorageProvider
//	storage.SetGauge(ctx, "value", Gauge(85.5))
//	if value, ok := storage.GetGauge(ctx, "value"); ok {
//		fmt.Printf("Value: %.1f%%", float64(value))
//	}
type StorageProvider interface {
	// GetGauge возвращает значение gauge метрики по ключу.
	// Возвращает значение и флаг существования метрики.
	GetGauge(ctx context.Context, key string) (Gauge, bool)

	// GetCounter возвращает значение counter метрики по ключу.
	// Возвращает значение и флаг существования метрики.
	GetCounter(ctx context.Context, key string) (Counter, bool)

	// KeysGauge возвращает список всех ключей gauge метрик.
	KeysGauge(ctx context.Context) []string

	// SetGauge устанавливает значение gauge метрики.
	SetGauge(ctx context.Context, key string, value Gauge)

	// SetCounter устанавливает значение counter метрики.
	SetCounter(ctx context.Context, key string, value Counter)

	// KeysCounter возвращает список всех ключей counter метрик.
	KeysCounter(ctx context.Context) []string

	// Ping проверяет доступность хранилища.
	// Возвращает ошибку, если хранилище недоступно.
	Ping(ctx context.Context) error
}

// Storage представляет универсальное хранилище для метрик типа T.
// Использует дженерики для типобезопасной работы с Gauge или Counter.
//
// Пример использования:
//
//	gaugeStorage := NewStorage[Gauge]()
//	gaugeStorage.Set("cpu_usage", Gauge(85.5))
//
//	counterStorage := NewStorage[Counter]()
//	counterStorage.Set("requests", Counter(1234))
type Storage[T interface{ Gauge | Counter }] struct {
	data map[string]T
}

// NewStorage создает новый экземпляр Storage для указанного типа метрики.
// Поддерживает типы Gauge и Counter.
//
// Пример использования:
//
//	// Создание хранилища для gauge метрик
//	gauges := NewStorage[Gauge]()
//
//	// Создание хранилища для counter метрик
//	counters := NewStorage[Counter]()
func NewStorage[T interface{ Gauge | Counter }]() *Storage[T] {
	return &Storage[T]{
		data: make(map[string]T),
	}
}

// Get возвращает значение метрики по ключу.
// Возвращает значение и флаг существования ключа в хранилище.
//
// Пример использования:
//
//	storage := NewStorage[Gauge]()
//	storage.Set("value", Gauge(23.5))
//	if value, ok := storage.Get("value"); ok {
//		fmt.Printf("Value: %.1f", float64(value))
//	}
func (s *Storage[T]) Get(key string) (T, bool) {
	val, ok := s.data[key]
	return val, ok
}

// Set устанавливает значение метрики по ключу.
//
// Пример использования:
//
//	storage := NewStorage[Counter]()
//	storage.Set("requests_total", Counter(1000))
func (s *Storage[T]) Set(key string, val T) {
	s.data[key] = val
}

// Keys возвращает срез всех ключей в хранилище.
// Порядок ключей не гарантируется.
//
// Пример использования:
//
//	storage := NewStorage[Gauge]()
//	storage.Set("cpu", Gauge(85.5))
//	storage.Set("memory", Gauge(67.2))
//	keys := storage.Keys() // ["cpu", "memory"] или ["memory", "cpu"]
func (s *Storage[T]) Keys() []string {
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

// Count увеличивает значение метрики на указанную величину.
// Если ключ не существует, создает новую запись со значением value.
//
// Пример использования:
//
//	storage := NewStorage[Counter]()
//	storage.Count("requests", Counter(1))  // Устанавливает значение 1
//	storage.Count("requests", Counter(5))  // Увеличивает до 6
func (s *Storage[T]) Count(key string, value T) {
	s.data[key] += value
}
