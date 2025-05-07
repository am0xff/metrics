package models

import storage "github.com/am0xff/metrics/internal/storagev2"

type Metrics struct {
	ID    string             `json:"id"`              // имя метрики
	MType storage.MetricType `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64             `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64           `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
