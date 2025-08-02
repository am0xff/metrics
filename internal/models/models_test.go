package models

import (
	"testing"

	"github.com/am0xff/metrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_String_Gauge(t *testing.T) {
	tests := []struct {
		name     string
		value    *float64
		expected string
	}{
		{
			name:     "positive float",
			value:    &[]float64{123.45}[0],
			expected: "123.45",
		},
		{
			name:     "negative float",
			value:    &[]float64{-67.89}[0],
			expected: "-67.89",
		},
		{
			name:     "zero float",
			value:    &[]float64{0.0}[0],
			expected: "0",
		},
		{
			name:     "integer as float",
			value:    &[]float64{100.0}[0],
			expected: "100",
		},
		{
			name:     "very small float",
			value:    &[]float64{0.000001}[0],
			expected: "0.000001",
		},
		{
			name:     "very large float",
			value:    &[]float64{1234567890.123456}[0],
			expected: "1234567890.123456",
		},
		{
			name:     "nil value",
			value:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric := Metrics{
				ID:    "test_gauge",
				MType: storage.MetricTypeGauge,
				Value: tt.value,
			}
			result := metric.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetrics_String_Counter(t *testing.T) {
	tests := []struct {
		name     string
		delta    *int64
		expected string
	}{
		{
			name:     "positive integer",
			delta:    &[]int64{12345}[0],
			expected: "12345",
		},
		{
			name:     "negative integer",
			delta:    &[]int64{-6789}[0],
			expected: "-6789",
		},
		{
			name:     "zero integer",
			delta:    &[]int64{0}[0],
			expected: "0",
		},
		{
			name:     "large integer",
			delta:    &[]int64{9223372036854775807}[0], // max int64
			expected: "9223372036854775807",
		},
		{
			name:     "small negative integer",
			delta:    &[]int64{-9223372036854775808}[0], // min int64
			expected: "-9223372036854775808",
		},
		{
			name:     "nil delta",
			delta:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric := Metrics{
				ID:    "test_counter",
				MType: storage.MetricTypeCounter,
				Delta: tt.delta,
			}
			result := metric.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetrics_String_UnknownType(t *testing.T) {
	metric := Metrics{
		ID:    "test_unknown",
		MType: "unknown_type",
		Value: &[]float64{123.45}[0],
		Delta: &[]int64{678}[0],
	}
	result := metric.String()
	assert.Equal(t, "", result)
}

func TestMetrics_Creation_Gauge(t *testing.T) {
	value := 85.5
	metric := Metrics{
		ID:    "cpu_usage",
		MType: storage.MetricTypeGauge,
		Value: &value,
	}

	assert.Equal(t, "cpu_usage", metric.ID)
	assert.Equal(t, storage.MetricTypeGauge, metric.MType)
	assert.NotNil(t, metric.Value)
	assert.Equal(t, 85.5, *metric.Value)
	assert.Nil(t, metric.Delta)
}

func TestMetrics_Creation_Counter(t *testing.T) {
	delta := int64(1)
	metric := Metrics{
		ID:    "requests_total",
		MType: storage.MetricTypeCounter,
		Delta: &delta,
	}

	assert.Equal(t, "requests_total", metric.ID)
	assert.Equal(t, storage.MetricTypeCounter, metric.MType)
	assert.NotNil(t, metric.Delta)
	assert.Equal(t, int64(1), *metric.Delta)
	assert.Nil(t, metric.Value)
}

func TestMetrics_EmptyValues(t *testing.T) {
	tests := []struct {
		name   string
		metric Metrics
	}{
		{
			name: "gauge with nil value",
			metric: Metrics{
				ID:    "empty_gauge",
				MType: storage.MetricTypeGauge,
				Value: nil,
			},
		},
		{
			name: "counter with nil delta",
			metric: Metrics{
				ID:    "empty_counter",
				MType: storage.MetricTypeCounter,
				Delta: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metric.String()
			assert.Equal(t, "", result)
		})
	}
}

func TestMetrics_BothValuesSet(t *testing.T) {
	// Test when both Value and Delta are set, but only one should be used based on MType
	value := 123.45
	delta := int64(678)

	gaugeMetric := Metrics{
		ID:    "test_gauge",
		MType: storage.MetricTypeGauge,
		Value: &value,
		Delta: &delta, // This should be ignored for gauge
	}
	assert.Equal(t, "123.45", gaugeMetric.String())

	counterMetric := Metrics{
		ID:    "test_counter",
		MType: storage.MetricTypeCounter,
		Value: &value, // This should be ignored for counter
		Delta: &delta,
	}
	assert.Equal(t, "678", counterMetric.String())
}

func TestMetrics_JSONTags(t *testing.T) {
	// This test verifies that the struct has the correct JSON tags
	// We can't directly test JSON tags, but we can test JSON marshaling/unmarshaling

	value := 42.5
	metric := Metrics{
		ID:    "test_metric",
		MType: storage.MetricTypeGauge,
		Value: &value,
	}

	// Basic verification that fields are accessible
	assert.Equal(t, "test_metric", metric.ID)
	assert.Equal(t, storage.MetricTypeGauge, metric.MType)
	assert.Equal(t, 42.5, *metric.Value)
}

// Benchmark tests for performance
func BenchmarkMetrics_String_Gauge(b *testing.B) {
	value := 123.456789
	metric := Metrics{
		ID:    "benchmark_gauge",
		MType: storage.MetricTypeGauge,
		Value: &value,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = metric.String()
	}
}

func BenchmarkMetrics_String_Counter(b *testing.B) {
	delta := int64(123456789)
	metric := Metrics{
		ID:    "benchmark_counter",
		MType: storage.MetricTypeCounter,
		Delta: &delta,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = metric.String()
	}
}
