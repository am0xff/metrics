package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	memstorage "github.com/am0xff/metrics/internal/storage/memory"
	"github.com/stretchr/testify/assert"
)

func TestSetupRoutes(t *testing.T) {
	storage := memstorage.NewStorage()
	handler := SetupRoutes(storage)

	assert.NotNil(t, handler)

	// Тест маршрутов
	server := httptest.NewServer(handler)
	defer server.Close()

	endpoints := []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/", 200},
		{"GET", "/ping", 200},
		{"GET", "/value/gauge/test", 404}, // метрика не существует
	}

	for _, ep := range endpoints {
		resp, err := http.Get(server.URL + ep.path)
		assert.NoError(t, err)
		assert.Equal(t, ep.status, resp.StatusCode)
		resp.Body.Close()
	}
}
