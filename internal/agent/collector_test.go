package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCollector(t *testing.T) {
	collector := NewCollector()

	assert.NotNil(t, collector)
	assert.Equal(t, int64(0), collector.pollCount)
}

func TestCollector_Collect(t *testing.T) {
	collector := NewCollector()

	gauges, counters := collector.Collect()

	// Проверяем, что возвращаются непустые карты
	assert.NotNil(t, gauges)
	assert.NotNil(t, counters)
	assert.NotEmpty(t, gauges)
	assert.NotEmpty(t, counters)

	// Проверяем наличие всех ожидаемых gauge метрик
	expectedGauges := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
		"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
		"Sys", "TotalAlloc", "RandomValue",
	}

	for _, metricName := range expectedGauges {
		value, exists := gauges[metricName]
		assert.True(t, exists, "Expected gauge metric %s not found", metricName)
		assert.GreaterOrEqual(t, value, 0.0, "Gauge metric %s should be non-negative", metricName)
	}

	// Проверяем, что RandomValue действительно случайное
	assert.GreaterOrEqual(t, gauges["RandomValue"], 0.0)
	assert.LessOrEqual(t, gauges["RandomValue"], 1.0)

	// Проверяем counter метрику PollCount
	pollCount, exists := counters["PollCount"]
	assert.True(t, exists, "Expected counter metric PollCount not found")
	assert.Equal(t, int64(1), pollCount, "PollCount should be 1 after first collection")

	// Проверяем, что у нас только одна counter метрика
	assert.Len(t, counters, 1, "Expected exactly one counter metric")
}

func TestCollector_CollectMultipleTimes(t *testing.T) {
	collector := NewCollector()

	// Первый сбор
	gauges1, counters1 := collector.Collect()
	pollCount1 := counters1["PollCount"]
	assert.Equal(t, int64(1), pollCount1)

	// Второй сбор
	gauges2, counters2 := collector.Collect()
	pollCount2 := counters2["PollCount"]
	assert.Equal(t, int64(2), pollCount2)

	// Третий сбор
	gauges3, counters3 := collector.Collect()
	pollCount3 := counters3["PollCount"]
	assert.Equal(t, int64(3), pollCount3)

	// Проверяем, что количество метрик остается постоянным
	assert.Len(t, gauges1, len(gauges2))
	assert.Len(t, gauges2, len(gauges3))
	assert.Len(t, counters1, len(counters2))
	assert.Len(t, counters2, len(counters3))

	// Проверяем, что RandomValue действительно изменяется (с высокой вероятностью)
	// Примечание: теоретически может быть одинаковым, но вероятность крайне мала
	randomValue1 := gauges1["RandomValue"]
	randomValue2 := gauges2["RandomValue"]
	randomValue3 := gauges3["RandomValue"]

	// Хотя бы одно из значений должно отличаться
	assert.True(t,
		randomValue1 != randomValue2 || randomValue2 != randomValue3 || randomValue1 != randomValue3,
		"RandomValue should change between collections")
}

func TestCollector_MetricsTypes(t *testing.T) {
	collector := NewCollector()
	gauges, counters := collector.Collect()

	// Проверяем типы всех gauge метрик
	for name, value := range gauges {
		assert.IsType(t, float64(0), value, "Gauge metric %s should be float64", name)
	}

	// Проверяем типы всех counter метрик
	for name, value := range counters {
		assert.IsType(t, int64(0), value, "Counter metric %s should be int64", name)
	}
}

func TestCollector_MemoryMetricsAreRealistic(t *testing.T) {
	collector := NewCollector()
	gauges, _ := collector.Collect()

	// Проверяем, что основные метрики памяти имеют разумные значения
	assert.Greater(t, gauges["Alloc"], 0.0, "Alloc should be greater than 0")
	assert.Greater(t, gauges["Sys"], 0.0, "Sys should be greater than 0")
	assert.Greater(t, gauges["TotalAlloc"], 0.0, "TotalAlloc should be greater than 0")

	// Проверяем, что HeapSys больше HeapInuse (логично для Go runtime)
	assert.GreaterOrEqual(t, gauges["HeapSys"], gauges["HeapInuse"],
		"HeapSys should be greater than or equal to HeapInuse")

	// Проверяем, что некоторые системные метрики имеют разумные значения
	assert.GreaterOrEqual(t, gauges["Sys"], gauges["HeapSys"],
		"Sys should be greater than or equal to HeapSys")
}

func TestCollector_PollCountIncrement(t *testing.T) {
	collector := NewCollector()

	// Проверяем начальное состояние
	assert.Equal(t, int64(0), collector.pollCount)

	// Собираем метрики несколько раз и проверяем инкремент
	for i := 1; i <= 5; i++ {
		_, counters := collector.Collect()

		// Проверяем внутреннее состояние
		assert.Equal(t, int64(i), collector.pollCount)

		// Проверяем возвращаемое значение
		assert.Equal(t, int64(i), counters["PollCount"])
	}
}

func TestCollector_RandomValueRange(t *testing.T) {
	collector := NewCollector()

	// Собираем метрики несколько раз и проверяем диапазон RandomValue
	for i := 0; i < 10; i++ {
		gauges, _ := collector.Collect()
		randomValue := gauges["RandomValue"]

		assert.GreaterOrEqual(t, randomValue, 0.0, "RandomValue should be >= 0")
		assert.LessOrEqual(t, randomValue, 1.0, "RandomValue should be <= 1")
	}
}

func TestCollector_ConsistentMetricNames(t *testing.T) {
	collector := NewCollector()

	// Собираем метрики дважды
	gauges1, counters1 := collector.Collect()
	gauges2, counters2 := collector.Collect()

	// Проверяем, что набор ключей одинаковый
	assert.Len(t, gauges1, len(gauges2), "Number of gauge metrics should be consistent")
	assert.Len(t, counters1, len(counters2), "Number of counter metrics should be consistent")

	// Проверяем, что все ключи присутствуют в обеих коллекциях
	for name := range gauges1 {
		_, exists := gauges2[name]
		assert.True(t, exists, "Gauge metric %s should exist in both collections", name)
	}

	for name := range counters1 {
		_, exists := counters2[name]
		assert.True(t, exists, "Counter metric %s should exist in both collections", name)
	}
}
