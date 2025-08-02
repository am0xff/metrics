package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCompressWriter(t *testing.T) {
	w := httptest.NewRecorder()
	key := "test-key"

	cw := newCompressWriter(w, key)

	assert.NotNil(t, cw)
	assert.Equal(t, w, cw.w)
	assert.NotNil(t, cw.zw)
	assert.Equal(t, key, cw.key)
}

func TestCompressWriter_Header(t *testing.T) {
	w := httptest.NewRecorder()
	cw := newCompressWriter(w, "")

	// Set a header and verify it's accessible
	expectedKey := "X-Test-Header"
	expectedValue := "test-value"
	cw.Header().Set(expectedKey, expectedValue)

	assert.Equal(t, expectedValue, cw.Header().Get(expectedKey))
	assert.Equal(t, expectedValue, w.Header().Get(expectedKey))
}

func TestCompressWriter_Write_WithKey(t *testing.T) {
	w := httptest.NewRecorder()
	key := "test-secret-key"
	cw := newCompressWriter(w, key)

	testData := []byte("test data for compression")

	n, err := cw.Write(testData)
	require.NoError(t, err)
	assert.Equal(t, len(testData), n)

	// Check that hash header is set
	hash := w.Header().Get("HashSHA256")
	assert.NotEmpty(t, hash)

	// Close to flush gzip data
	err = cw.Close()
	require.NoError(t, err)
}

func TestCompressWriter_Write_WithoutKey(t *testing.T) {
	w := httptest.NewRecorder()
	cw := newCompressWriter(w, "") // empty key

	testData := []byte("test data without key")

	n, err := cw.Write(testData)
	require.NoError(t, err)
	assert.Equal(t, len(testData), n)

	// Check that hash header is NOT set
	hash := w.Header().Get("HashSHA256")
	assert.Empty(t, hash)
}

func TestCompressWriter_WriteHeader_JSON(t *testing.T) {
	w := httptest.NewRecorder()
	cw := newCompressWriter(w, "")

	// Set JSON content type
	cw.Header().Set("Content-Type", "application/json")

	// Write success status
	cw.WriteHeader(http.StatusOK)

	// Check that gzip encoding is set
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCompressWriter_WriteHeader_HTML(t *testing.T) {
	w := httptest.NewRecorder()
	cw := newCompressWriter(w, "")

	// Set HTML content type
	cw.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Write success status
	cw.WriteHeader(http.StatusOK)

	// Check that gzip encoding is set
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
}

func TestCompressWriter_WriteHeader_NonCompressible(t *testing.T) {
	w := httptest.NewRecorder()
	cw := newCompressWriter(w, "")

	// Set non-compressible content type
	cw.Header().Set("Content-Type", "image/png")

	// Write success status
	cw.WriteHeader(http.StatusOK)

	// Check that gzip encoding is NOT set
	assert.Empty(t, w.Header().Get("Content-Encoding"))
}

func TestCompressWriter_WriteHeader_ErrorStatus(t *testing.T) {
	w := httptest.NewRecorder()
	cw := newCompressWriter(w, "")

	// Set JSON content type
	cw.Header().Set("Content-Type", "application/json")

	// Write error status
	cw.WriteHeader(http.StatusInternalServerError)

	// Check that gzip encoding is NOT set for error status
	assert.Empty(t, w.Header().Get("Content-Encoding"))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCompressWriter_Close(t *testing.T) {
	w := httptest.NewRecorder()
	cw := newCompressWriter(w, "")

	// Write some data
	_, err := cw.Write([]byte("test data"))
	require.NoError(t, err)

	// Close should work without error
	err = cw.Close()
	assert.NoError(t, err)

	// Second close should also work (idempotent)
	err = cw.Close()
	assert.NoError(t, err)
}

func TestCompressWriter_Close_NilWriter(t *testing.T) {
	cw := &compressWriter{zw: nil}

	err := cw.Close()
	assert.NoError(t, err)
}

func TestNewCompressReader_Success(t *testing.T) {
	// Create gzipped data
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	testData := "test data for decompression"
	_, err := gw.Write([]byte(testData))
	require.NoError(t, err)
	err = gw.Close()
	require.NoError(t, err)

	// Create reader from gzipped data
	reader := io.NopCloser(bytes.NewReader(buf.Bytes()))
	cr, err := newCompressReader(reader)

	require.NoError(t, err)
	assert.NotNil(t, cr)
	assert.NotNil(t, cr.r)
	assert.NotNil(t, cr.zr)
}

func TestNewCompressReader_InvalidGzip(t *testing.T) {
	// Create invalid gzip data
	invalidData := []byte("this is not gzipped data")
	reader := io.NopCloser(bytes.NewReader(invalidData))

	cr, err := newCompressReader(reader)

	assert.Error(t, err)
	assert.Nil(t, cr)
}

func TestCompressReader_Read(t *testing.T) {
	// Create gzipped data
	testData := "Hello, this is test data for compression and decompression!"
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte(testData))
	require.NoError(t, err)
	err = gw.Close()
	require.NoError(t, err)

	// Create compress reader
	reader := io.NopCloser(bytes.NewReader(buf.Bytes()))
	cr, err := newCompressReader(reader)
	require.NoError(t, err)

	// Read and verify data
	result, err := io.ReadAll(cr)
	require.NoError(t, err)
	assert.Equal(t, testData, string(result))
}

func TestCompressReader_Close(t *testing.T) {
	// Create gzipped data
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte("test"))
	require.NoError(t, err)
	err = gw.Close()
	require.NoError(t, err)

	// Create compress reader
	reader := io.NopCloser(bytes.NewReader(buf.Bytes()))
	cr, err := newCompressReader(reader)
	require.NoError(t, err)

	// Close should work without error
	err = cr.Close()
	assert.NoError(t, err)
}

func TestCompressReader_Close_NilReaders(t *testing.T) {
	cr := &compressReader{r: nil, zr: nil}

	err := cr.Close()
	assert.NoError(t, err)
}

func TestGzipMiddleware_NoCompression(t *testing.T) {
	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	// Create middleware
	middleware := GzipMiddleware(testHandler, "")

	// Create request without compression headers
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Execute
	middleware.ServeHTTP(w, req)

	// Verify no compression applied
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello, World!", w.Body.String())
	assert.Empty(t, w.Header().Get("Content-Encoding"))
}

func TestGzipMiddleware_WithCompression(t *testing.T) {
	// Create test handler that sets JSON content type
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Hello, World!"}`))
	})

	// Create middleware
	middleware := GzipMiddleware(testHandler, "test-key")

	// Create request with gzip acceptance
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	// Execute
	middleware.ServeHTTP(w, req)

	// Verify compression applied
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	assert.NotEmpty(t, w.Header().Get("HashSHA256"))

	// Verify compressed data can be decompressed
	gr, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.Equal(t, `{"message": "Hello, World!"}`, string(decompressed))
}

func TestGzipMiddleware_WithDecompression(t *testing.T) {
	// Create test handler that reads request body
	var receivedBody string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	})

	// Create compressed request body
	testData := `{"key": "value", "number": 42}`
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte(testData))
	require.NoError(t, err)
	err = gw.Close()
	require.NoError(t, err)

	// Create middleware
	middleware := GzipMiddleware(testHandler, "")

	// Create request with compressed body
	req := httptest.NewRequest("POST", "/test", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()

	// Execute
	middleware.ServeHTTP(w, req)

	// Verify decompression worked
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, testData, receivedBody)
}

func TestGzipMiddleware_InvalidCompressedRequest(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware
	middleware := GzipMiddleware(testHandler, "")

	// Create request with invalid compressed data
	invalidData := strings.NewReader("invalid gzip data")
	req := httptest.NewRequest("POST", "/test", invalidData)
	req.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()

	// Execute
	middleware.ServeHTTP(w, req)

	// Should return internal server error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGzipMiddleware_BothCompressionAndDecompression(t *testing.T) {
	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"received": "` + string(body) + `"}`))
	})

	// Create compressed request body
	requestData := "test request data"
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte(requestData))
	require.NoError(t, err)
	err = gw.Close()
	require.NoError(t, err)

	// Create middleware
	middleware := GzipMiddleware(testHandler, "secret")

	// Create request with both compression headers
	req := httptest.NewRequest("POST", "/test", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	// Execute
	middleware.ServeHTTP(w, req)

	// Verify both compression and decompression worked
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	assert.NotEmpty(t, w.Header().Get("HashSHA256"))

	// Decompress response and verify
	gr, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.Equal(t, `{"received": "test request data"}`, string(decompressed))
}
