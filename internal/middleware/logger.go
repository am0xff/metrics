package middleware

import (
	"net/http"
	"time"

	"github.com/am0xff/metrics/internal/logger"
	"go.uber.org/zap"
)

func LoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &logger.ResponseData{
			Status: 0,
			Size:   0,
		}

		lw := logger.LoggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			ResponseData:   responseData,
		}

		uri := r.RequestURI
		method := r.Method

		// call handler
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		// Сведения о запросах должны содержать URI, метод запроса и время, затраченное на его выполнение.
		// Сведения об ответах должны содержать код статуса и размер содержимого ответа.
		logger.Log.Info("request",
			zap.String("uri", uri),
			zap.String("method", method),
			zap.Duration("duration", duration),
			zap.Int("status", responseData.Status),
			zap.Int("size", responseData.Size),
		)
	})
}
