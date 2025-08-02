package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/am0xff/metrics/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashMiddleware_NoValidation(t *testing.T) {
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := HashMiddleware(testHandler, "")

	// Without key - no validation
	req := httptest.NewRequest("POST", "/update/", strings.NewReader("test"))
	req.Header.Set("HashSHA256", "hash")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHashMiddleware_SkipValidation(t *testing.T) {
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := HashMiddleware(testHandler, "key")

	// GET method - skip validation
	req := httptest.NewRequest("GET", "/update/", nil)
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)
	assert.True(t, handlerCalled)

	// Wrong path - skip validation
	handlerCalled = false
	req = httptest.NewRequest("POST", "/other/", nil)
	w = httptest.NewRecorder()
	middleware.ServeHTTP(w, req)
	assert.True(t, handlerCalled)

	// No hash header - skip validation
	handlerCalled = false
	req = httptest.NewRequest("POST", "/update/", strings.NewReader("test"))
	w = httptest.NewRecorder()
	middleware.ServeHTTP(w, req)
	assert.True(t, handlerCalled)
}

func TestHashMiddleware_ValidHash(t *testing.T) {
	var receivedBody string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	})

	key := "secret-key"
	middleware := HashMiddleware(testHandler, key)

	testData := "test data"
	validHash := utils.CreateHash([]byte(testData), key)

	req := httptest.NewRequest("POST", "/update/", strings.NewReader(testData))
	req.Header.Set("HashSHA256", validHash)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, testData, receivedBody)
}

func TestHashMiddleware_InvalidHash(t *testing.T) {
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	middleware := HashMiddleware(testHandler, "secret-key")

	req := httptest.NewRequest("POST", "/update/", strings.NewReader("test data"))
	req.Header.Set("HashSHA256", "invalid-hash")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHashMiddleware_ReadError(t *testing.T) {
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	middleware := HashMiddleware(testHandler, "key")

	// Use error reader
	errorReader := &errorReader{}
	req := httptest.NewRequest("POST", "/update/", errorReader)
	req.Header.Set("HashSHA256", "hash")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}
