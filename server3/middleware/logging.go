package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"api-ondemand/server3/metrics"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func wrapWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(m *metrics.Metrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := wrapWriter(w)
		next.ServeHTTP(rw, r)
		latencyMs := float64(time.Since(start).Microseconds()) / 1000.0
		m.Record(r.URL.Path, latencyMs)

		entry := map[string]any{
			"time":       time.Now().UTC().Format(time.RFC3339),
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     rw.statusCode,
			"latency_ms": latencyMs,
			"remote_ip":  r.RemoteAddr,
		}
		if b, err := json.Marshal(entry); err == nil {
			log.Println(string(b))
		}
	})
}
