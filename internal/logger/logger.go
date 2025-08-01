package logger

import (
	"net/http"

	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

func Initialize() error {
	lvl, err := zap.ParseAtomicLevel("info")
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl

	return nil
}

type (
	ResponseData struct {
		Status int
		Size   int
	}

	LoggingResponseWriter struct {
		http.ResponseWriter
		ResponseData *ResponseData
	}
)

func (r *LoggingResponseWriter) Write(data []byte) (int, error) {
	size, err := r.ResponseWriter.Write(data)
	r.ResponseData.Size += size
	return size, err
}

func (r *LoggingResponseWriter) WriteHeader(status int) {
	r.ResponseWriter.WriteHeader(status)
	r.ResponseData.Status = status
}
