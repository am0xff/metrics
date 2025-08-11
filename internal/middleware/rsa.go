package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/am0xff/metrics/internal/utils"
)

func RSAMiddleware(next http.Handler, cryptoKeyPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Если путь к ключу не указан, пропускаем без изменений
		if cryptoKeyPath == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Проверяем, что это POST запрос к нужным эндпоинтам
		if r.Method != http.MethodPost ||
			(r.URL.Path != "/update/" && r.URL.Path != "/updates/") {
			next.ServeHTTP(w, r)
			return
		}

		// Определяем зашифрованные данные по отсутствию Content-Encoding: gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		if contentEncoding == "gzip" {
			// Данные только сжаты, не зашифрованы
			next.ServeHTTP(w, r)
			return
		}

		// Читаем тело запроса
		encryptedData, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("RSA middleware: failed to read request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		r.Body.Close()

		// Загружаем приватный ключ
		privateKey, err := utils.LoadPrivateKey(cryptoKeyPath)
		if err != nil {
			log.Printf("RSA middleware: failed to load private key: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Расшифровываем данные
		decryptedData, err := utils.DecryptRSA(encryptedData, privateKey)
		if err != nil {
			log.Printf("RSA middleware: failed to decrypt data: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Заменяем тело запроса расшифрованными данными
		r.Body = io.NopCloser(bytes.NewReader(decryptedData))

		// Устанавливаем Content-Encoding: gzip для дальнейшей обработки
		// (так как агент сначала сжимает, а потом шифрует)
		r.Header.Set("Content-Encoding", "gzip")

		// Передаем управление следующему middleware
		next.ServeHTTP(w, r)
	})
}
