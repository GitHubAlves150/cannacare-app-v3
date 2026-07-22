// ================================================================
// PACOTE MIDDLEWARE - LOGGER
// ================================================================
// Middleware para logging de requisições.
// ================================================================

package middleware

import (
	"log"
	"net/http"
	"time"
)

// ================================================================
// MIDDLEWARE: LOGGERMIDDLEWARE
// ================================================================
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer para capturar status code
		wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		// Processar requisição
		next.ServeHTTP(wrapped, r)

		// Log da requisição
		log.Printf("📝 %s %s %d %s",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			time.Since(start),
		)
	})
}

// ================================================================
// RESPONSE WRITER WRAPPER
// ================================================================
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
