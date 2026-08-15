package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

const (
	defaultPort       = "8080"
	defaultServerName = "Server 1"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ---------------------------------------------------------------------------
// Metrics
// ---------------------------------------------------------------------------

type metrics struct {
	mu              sync.Mutex
	totalRequests   int64
	requestsByPath  map[string]int64
	totalLatencyMs  float64
}

func newMetrics() *metrics {
	return &metrics{requestsByPath: make(map[string]int64)}
}

func (m *metrics) record(path string, latencyMs float64) {
	atomic.AddInt64(&m.totalRequests, 1)
	m.mu.Lock()
	m.requestsByPath[path]++
	m.totalLatencyMs += latencyMs
	m.mu.Unlock()
}

func (m *metrics) snapshot() (total int64, byPath map[string]int64, avgLatency float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	total = atomic.LoadInt64(&m.totalRequests)
	byPath = make(map[string]int64, len(m.requestsByPath))
	for k, v := range m.requestsByPath {
		byPath[k] = v
	}
	if total > 0 {
		avgLatency = m.totalLatencyMs / float64(total)
	}
	return
}

// ---------------------------------------------------------------------------
// Structured JSON logger middleware
// ---------------------------------------------------------------------------

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

func loggingMiddleware(m *metrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := wrapWriter(w)
		next.ServeHTTP(rw, r)
		latencyMs := float64(time.Since(start).Microseconds()) / 1000.0
		m.record(r.URL.Path, latencyMs)

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

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf(`{"error":"failed to encode response","detail":"%v"}`, err)
	}
}

func rootHandler(serverName, port string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w,
			`<!DOCTYPE html><html><body><h1>👋 Hello from %s</h1><p>Listening on port <strong>%s</strong></p></body></html>`,
			serverName, port,
		)
	}
}

func pingHandler(serverName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "server": serverName})
	}
}

func infoHandler(serverName, port string, startTime time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostname, _ := os.Hostname()
		writeJSON(w, http.StatusOK, map[string]any{
			"name":       serverName,
			"port":       port,
			"uptime":     time.Since(startTime).String(),
			"go_version": runtime.Version(),
			"hostname":   hostname,
		})
	}
}

func healthHandler(serverName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"healthy": true, "server": serverName})
	}
}

func echoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MB cap
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read body"})
			return
		}
		defer r.Body.Close()
		writeJSON(w, http.StatusOK, map[string]string{"echo": string(body)})
	}
}

func metricsHandler(m *metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		total, byPath, avgLatency := m.snapshot()
		writeJSON(w, http.StatusOK, map[string]any{
			"total_requests":    total,
			"requests_by_path":  byPath,
			"avg_latency_ms":    avgLatency,
		})
	}
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	port := getEnv("PORT", defaultPort)
	serverName := getEnv("SERVER_NAME", defaultServerName)
	startTime := time.Now()

	m := newMetrics()

	mux := http.NewServeMux()
	mux.Handle("/", rootHandler(serverName, port))
	mux.Handle("/ping", pingHandler(serverName))
	mux.Handle("/info", infoHandler(serverName, port, startTime))
	mux.Handle("/health", healthHandler(serverName))
	mux.Handle("/echo", echoHandler())
	mux.Handle("/metrics", metricsHandler(m))

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      loggingMiddleware(m, mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background.
	go func() {
		log.Printf(`{"event":"starting","server":"%s","port":"%s"}`, serverName, port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf(`{"event":"fatal","error":"%v"}`, err)
		}
	}()

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf(`{"event":"shutting_down","server":"%s"}`, serverName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf(`{"event":"shutdown_error","error":"%v"}`, err)
	}
	log.Printf(`{"event":"stopped","server":"%s"}`, serverName)
}
