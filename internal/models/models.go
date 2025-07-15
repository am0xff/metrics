// Package models предоставляет типы данных для работы с метриками.
// Пакет содержит основные структуры для передачи и обработки метрик
// типов gauge и counter в системе сбора метрик.
package models

import (
	"strconv"

	"github.com/am0xff/metrics/internal/storage"
)

// Metrics представляет структуру данных для метрики.
// Структура поддерживает два типа метрик: gauge и counter.
// Для gauge используется поле Value, для counter - поле Delta.
//
// Пример использования:
//
//	// Создание gauge метрики
//	gaugeMetric := Metrics{
//		ID:    "cpu_usage",
//		MType: storage.MetricTypeGauge,
//		Value: &[]float64{85.5}[0],
//	}
//
//	// Создание counter метрики
//	counterMetric := Metrics{
//		ID:    "requests_total",
//		MType: storage.MetricTypeCounter,
//		Delta: &[]int64{1}[0],
//	}
type Metrics struct {
	ID    string             `json:"id"`              // имя метрики
	MType storage.MetricType `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64             `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64           `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// String возвращает строковое представление значения метрики.
// Для gauge возвращает отформатированное значение float64,
// для counter - значение int64.
// Если значение не установлено, возвращает пустую строку.
//
// Пример использования:
//
//	metric := Metrics{
//		ID:    "temperature",
//		MType: storage.MetricTypeGauge,
//		Value: &[]float64{23.5}[0],
//	}
//	fmt.Println(metric.String()) // Выведет: "23.5"
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
