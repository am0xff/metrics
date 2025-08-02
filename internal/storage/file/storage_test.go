package file

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/am0xff/metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStorage_NoRestore(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		Restore:         false,
		FileStoragePath: "/tmp/test.json",
		StoreInterval:   1,
	}

	fs, err := NewStorage(ctx, cfg)

	require.NoError(t, err)
	assert.NotNil(t, fs)
	assert.NotNil(t, fs.ms)
	assert.Equal(t, cfg, fs.cfg)
}

func TestNewStorage_RestoreFileNotExists(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		Restore:         true,
		FileStoragePath: "/tmp/nonexistent.json",
		StoreInterval:   1,
	}

	fs, err := NewStorage(ctx, cfg)

	require.NoError(t, err)
	assert.NotNil(t, fs)
}

func TestNewStorage_RestoreWithData(t *testing.T) {
	ctx := context.Background()

	// Create temp file with test data
	tempFile := filepath.Join(t.TempDir(), "test.json")
	testData := DumpStorage{
		Gauges:   map[string]storage.Gauge{"gauge1": 123.45},
		Counters: map[string]storage.Counter{"counter1": 100},
	}
	data, err := json.Marshal(testData)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(tempFile, data, 0644))

	cfg := Config{
		Restore:         true,
		FileStoragePath: tempFile,
		StoreInterval:   1,
	}

	fs, err := NewStorage(ctx, cfg)

	require.NoError(t, err)

	// Verify data was loaded
	gauge, exists := fs.GetGauge(ctx, "gauge1")
	assert.True(t, exists)
	assert.Equal(t, storage.Gauge(123.45), gauge)

	counter, exists := fs.GetCounter(ctx, "counter1")
	assert.True(t, exists)
	assert.Equal(t, storage.Counter(100), counter)
}

func TestNewStorage_RestoreInvalidJSON(t *testing.T) {
	ctx := context.Background()

	// Create temp file with invalid JSON
	tempFile := filepath.Join(t.TempDir(), "invalid.json")
	require.NoError(t, os.WriteFile(tempFile, []byte("invalid json"), 0644))

	cfg := Config{
		Restore:         true,
		FileStoragePath: tempFile,
		StoreInterval:   1,
	}

	fs, err := NewStorage(ctx, cfg)

	assert.Error(t, err)
	assert.Nil(t, fs)
}

func TestFileStorage_SetAndGetGauge(t *testing.T) {
	ctx := context.Background()
	cfg := Config{StoreInterval: 1} // Don't auto-save

	fs, err := NewStorage(ctx, cfg)
	require.NoError(t, err)

	// Set gauge
	fs.SetGauge(ctx, "test_gauge", storage.Gauge(42.5))

	// Get gauge
	value, exists := fs.GetGauge(ctx, "test_gauge")
	assert.True(t, exists)
	assert.Equal(t, storage.Gauge(42.5), value)

	// Get non-existing gauge
	_, exists = fs.GetGauge(ctx, "nonexistent")
	assert.False(t, exists)
}

func TestFileStorage_SetAndGetCounter(t *testing.T) {
	ctx := context.Background()
	cfg := Config{StoreInterval: 1} // Don't auto-save

	fs, err := NewStorage(ctx, cfg)
	require.NoError(t, err)

	// Set counter
	fs.SetCounter(ctx, "test_counter", storage.Counter(150))

	// Get counter
	value, exists := fs.GetCounter(ctx, "test_counter")
	assert.True(t, exists)
	assert.Equal(t, storage.Counter(150), value)

	// Get non-existing counter
	_, exists = fs.GetCounter(ctx, "nonexistent")
	assert.False(t, exists)
}

func TestFileStorage_Keys(t *testing.T) {
	ctx := context.Background()
	cfg := Config{StoreInterval: 1}

	fs, err := NewStorage(ctx, cfg)
	require.NoError(t, err)

	// Add some data
	fs.SetGauge(ctx, "gauge1", storage.Gauge(1.1))
	fs.SetGauge(ctx, "gauge2", storage.Gauge(2.2))
	fs.SetCounter(ctx, "counter1", storage.Counter(10))
	fs.SetCounter(ctx, "counter2", storage.Counter(20))

	// Check gauge keys
	gaugeKeys := fs.KeysGauge(ctx)
	assert.Len(t, gaugeKeys, 2)
	assert.Contains(t, gaugeKeys, "gauge1")
	assert.Contains(t, gaugeKeys, "gauge2")

	// Check counter keys
	counterKeys := fs.KeysCounter(ctx)
	assert.Len(t, counterKeys, 2)
	assert.Contains(t, counterKeys, "counter1")
	assert.Contains(t, counterKeys, "counter2")
}

func TestFileStorage_AutoSave(t *testing.T) {
	ctx := context.Background()
	tempFile := filepath.Join(t.TempDir(), "autosave.json")

	cfg := Config{
		FileStoragePath: tempFile,
		StoreInterval:   0, // Auto-save enabled
	}

	fs, err := NewStorage(ctx, cfg)
	require.NoError(t, err)

	// Set data (should trigger auto-save)
	fs.SetGauge(ctx, "auto_gauge", storage.Gauge(99.9))
	fs.SetCounter(ctx, "auto_counter", storage.Counter(999))

	// Give it a moment for file operations
	time.Sleep(10 * time.Millisecond)

	// Verify file was created and contains data
	assert.FileExists(t, tempFile)

	// Read and verify file content
	data, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	var dump DumpStorage
	require.NoError(t, json.Unmarshal(data, &dump))

	assert.Equal(t, storage.Gauge(99.9), dump.Gauges["auto_gauge"])
	assert.Equal(t, storage.Counter(999), dump.Counters["auto_counter"])
}

func TestFileStorage_MarshalJSON(t *testing.T) {
	ctx := context.Background()
	cfg := Config{StoreInterval: 1}

	fs, err := NewStorage(ctx, cfg)
	require.NoError(t, err)

	// Add test data
	fs.SetGauge(ctx, "gauge1", storage.Gauge(123.45))
	fs.SetCounter(ctx, "counter1", storage.Counter(678))

	// Marshal to JSON
	data, err := fs.MarshalJSON()
	require.NoError(t, err)

	// Unmarshal and verify
	var dump DumpStorage
	require.NoError(t, json.Unmarshal(data, &dump))

	assert.Equal(t, storage.Gauge(123.45), dump.Gauges["gauge1"])
	assert.Equal(t, storage.Counter(678), dump.Counters["counter1"])
}

func TestFileStorage_Save(t *testing.T) {
	ctx := context.Background()
	tempFile := filepath.Join(t.TempDir(), "save_test.json")

	cfg := Config{
		FileStoragePath: tempFile,
		StoreInterval:   1, // Manual save
	}

	fs, err := NewStorage(ctx, cfg)
	require.NoError(t, err)

	// Add data
	fs.SetGauge(ctx, "save_gauge", storage.Gauge(456.78))
	fs.SetCounter(ctx, "save_counter", storage.Counter(321))

	// Manual save
	err = fs.Save()
	require.NoError(t, err)

	// Verify file exists and content
	assert.FileExists(t, tempFile)

	data, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	var dump DumpStorage
	require.NoError(t, json.Unmarshal(data, &dump))

	assert.Equal(t, storage.Gauge(456.78), dump.Gauges["save_gauge"])
	assert.Equal(t, storage.Counter(321), dump.Counters["save_counter"])
}

func TestFileStorage_Ping(t *testing.T) {
	ctx := context.Background()
	cfg := Config{StoreInterval: 1}

	fs, err := NewStorage(ctx, cfg)
	require.NoError(t, err)

	// Ping should always return nil for file storage
	err = fs.Ping(ctx)
	assert.NoError(t, err)
}
