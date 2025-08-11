package memory

import (
	"context"
	"testing"

	"github.com/am0xff/metrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewStorage(t *testing.T) {
	store := NewStorage()

	assert.NotNil(t, store)
	assert.NotNil(t, store.Gauges)
	assert.NotNil(t, store.Counters)
}

func TestMemStorage_Gauge_Operations(t *testing.T) {
	store := NewStorage()
	ctx := context.Background()

	// Test setting and getting gauge
	testKey := "test_gauge"
	testValue := storage.Gauge(123.45)

	// Set gauge
	store.SetGauge(ctx, testKey, testValue)

	// Get gauge
	value, exists := store.GetGauge(ctx, testKey)
	assert.True(t, exists)
	assert.Equal(t, testValue, value)

	// Test non-existing gauge
	_, exists = store.GetGauge(ctx, "non_existing")
	assert.False(t, exists)
}

func TestMemStorage_Counter_Operations(t *testing.T) {
	store := NewStorage()
	ctx := context.Background()

	// Test setting and getting counter
	testKey := "test_counter"
	testValue := storage.Counter(100)

	// Set counter
	store.SetCounter(ctx, testKey, testValue)

	// Get counter
	value, exists := store.GetCounter(ctx, testKey)
	assert.True(t, exists)
	assert.Equal(t, testValue, value)

	// Test counter accumulation (if Count method adds values)
	store.SetCounter(ctx, testKey, storage.Counter(50))
	value, exists = store.GetCounter(ctx, testKey)
	assert.True(t, exists)
	// Assuming Count method accumulates values
	assert.Equal(t, storage.Counter(150), value)

	// Test non-existing counter
	_, exists = store.GetCounter(ctx, "non_existing")
	assert.False(t, exists)
}

func TestMemStorage_Keys(t *testing.T) {
	store := NewStorage()
	ctx := context.Background()

	// Initially no keys
	gaugeKeys := store.KeysGauge(ctx)
	counterKeys := store.KeysCounter(ctx)
	assert.Empty(t, gaugeKeys)
	assert.Empty(t, counterKeys)

	// Add some gauges and counters
	store.SetGauge(ctx, "gauge1", storage.Gauge(1.1))
	store.SetGauge(ctx, "gauge2", storage.Gauge(2.2))
	store.SetCounter(ctx, "counter1", storage.Counter(10))
	store.SetCounter(ctx, "counter2", storage.Counter(20))

	// Check keys
	gaugeKeys = store.KeysGauge(ctx)
	counterKeys = store.KeysCounter(ctx)

	assert.Len(t, gaugeKeys, 2)
	assert.Contains(t, gaugeKeys, "gauge1")
	assert.Contains(t, gaugeKeys, "gauge2")

	assert.Len(t, counterKeys, 2)
	assert.Contains(t, counterKeys, "counter1")
	assert.Contains(t, counterKeys, "counter2")
}

func TestMemStorage_Ping(t *testing.T) {
	store := NewStorage()
	ctx := context.Background()

	// Ping should always return nil for memory storage
	err := store.Ping(ctx)
	assert.NoError(t, err)
}

func TestMemStorage_Multiple_Operations(t *testing.T) {
	store := NewStorage()
	ctx := context.Background()

	// Test multiple gauges
	gauges := map[string]storage.Gauge{
		"cpu_usage":    storage.Gauge(75.5),
		"memory_usage": storage.Gauge(60.2),
		"disk_usage":   storage.Gauge(45.8),
	}

	for key, value := range gauges {
		store.SetGauge(ctx, key, value)
	}

	// Verify all gauges
	for key, expectedValue := range gauges {
		value, exists := store.GetGauge(ctx, key)
		assert.True(t, exists, "Gauge %s should exist", key)
		assert.Equal(t, expectedValue, value, "Gauge %s value mismatch", key)
	}

	// Test multiple counters
	counters := map[string]storage.Counter{
		"requests":  storage.Counter(1000),
		"errors":    storage.Counter(50),
		"successes": storage.Counter(950),
	}

	for key, value := range counters {
		store.SetCounter(ctx, key, value)
	}

	// Verify all counters
	for key, expectedValue := range counters {
		value, exists := store.GetCounter(ctx, key)
		assert.True(t, exists, "Counter %s should exist", key)
		assert.Equal(t, expectedValue, value, "Counter %s value mismatch", key)
	}
}
