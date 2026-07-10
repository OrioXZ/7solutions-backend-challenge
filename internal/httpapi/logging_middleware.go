package httpapi

import (
	"log/slog"
	"net/http"
	"time"
)

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(statusCode int) {
	if w.statusCode != 0 {
		return
	}
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusResponseWriter) Write(body []byte) (int, error) {
	if w.statusCode == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

// LoggingMiddleware records one structured access log after each HTTP request.
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			responseWriter := &statusResponseWriter{ResponseWriter: w}

			next.ServeHTTP(responseWriter, r)

			statusCode := responseWriter.statusCode
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			logger.InfoContext(
				r.Context(),
				"HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", statusCode,
				"duration_ms", time.Since(startedAt).Milliseconds(),
			)
		})
	}
}
