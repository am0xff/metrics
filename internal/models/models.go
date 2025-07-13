package models

import (
	"strconv"

	"github.com/am0xff/metrics/internal/storage"
)

type Metrics struct {
	ID    string             `json:"id"`              // имя метрики
	MType storage.MetricType `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64             `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64           `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m Metrics) String() string {
	switch m.MType {
	case storage.MetricTypeGauge:
		if m.Value != nil {
			return strconv.FormatFloat(*m.Value, 'f', -1, 64)
		}
	case storage.MetricTypeCounter:
		if m.Delta != nil {
			return strconv.FormatInt(*m.Delta, 10)
		}
	}
	return ""
}
