// Package router предоставляет настройку маршрутизации для API сервиса метрик.
// Пакет содержит функции для создания и конфигурации HTTP маршрутов,
// связывающих URL эндпоинты с соответствующими обработчиками.
package router

import (
	"net/http"

	"github.com/am0xff/metrics/internal/handlers"
	"github.com/am0xff/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

// SetupRoutes создает и настраивает HTTP маршрутизатор для API метрик.
// Принимает провайдер хранилища и возвращает настроенный HTTP обработчик
// со всеми необходимыми маршрутами.
//
// Настроенные маршруты:
//
//	GET  /                              - HTML страница со всеми метриками
//	GET  /ping                          - проверка доступности хранилища
//	POST /value/                        - получение метрики (JSON)
//	POST /update/                       - обновление метрики (JSON)
//	POST /updates/                      - массовое обновление метрик (JSON)
//	GET  /value/{type}/{name}           - получение метрики (URL параметры)
//	POST /update/{type}/{name}/{value}  - обновление метрики (URL параметры)
//
// Параметры маршрутов:
//   - {type}: тип метрики ("gauge" или "counter")
//   - {name}: имя метрики
//   - {value}: значение метрики
//
// Пример использования:
//
//	// Создание хранилища
//	storage := memory.NewMemoryStorage()
//
//	// Настройка маршрутов
//	handler := SetupRoutes(storage)
//
//	// Запуск HTTP сервера
//	log.Fatal(http.ListenAndServe(":8080", handler))
//
// Примеры запросов:
//
//	# Получение всех метрик в HTML
//	curl http://localhost:8080/
//
//	# Проверка доступности
//	curl http://localhost:8080/ping
//
//	# Обновление gauge метрики через URL
//	curl -X POST http://localhost:8080/update/gauge/cpu_usage/85.5
//
//	# Получение counter метрики через URL
//	curl http://localhost:8080/value/counter/requests_total
//
//	# Обновление метрики через JSON
//	curl -X POST http://localhost:8080/update/ \
//		-H "Content-Type: application/json" \
//		-d '{"id":"memory_usage","type":"gauge","value":67.2}'
//
//	# Массовое обновление через JSON
//	curl -X POST http://localhost:8080/updates/ \
//		-H "Content-Type: application/json" \
//		-d '[{"id":"cpu","type":"gauge","value":85.5},{"id":"requests","type":"counter","delta":100}]'
func SetupRoutes(sp storage.StorageProvider) http.Handler {
	r := chi.NewRouter()

	handler := handlers.NewHandler(sp)

	r.Get("/", handler.GetMetrics)
	r.Get("/ping", handler.Ping)
	r.Post("/value/", handler.POSTGetMetric)
	r.Post("/update/", handler.POSTUpdateMetric)
	r.Post("/updates/", handler.POSTUpdatesMetrics)
	r.Get("/value/{type}/{name}", handler.GETGetMetric)
	r.Post("/update/{type}/{name}/{value}", handler.GETUpdateMetric)
	return r
}
