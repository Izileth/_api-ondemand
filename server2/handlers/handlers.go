package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"api-ondemand/server2/metrics"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf(`{"error":"failed to encode response","detail":"%v"}`, err)
	}
}

func RootHandler(serverName, port string) http.HandlerFunc {
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

func PingHandler(serverName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "server": serverName})
	}
}

func InfoHandler(serverName, port string, startTime time.Time) http.HandlerFunc {
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

func HealthHandler(serverName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"healthy": true, "server": serverName})
	}
}

func EchoHandler() http.HandlerFunc {
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

func MetricsHandler(m *metrics.Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		total, byPath, avgLatency := m.Snapshot()
		writeJSON(w, http.StatusOK, map[string]any{
			"total_requests":   total,
			"requests_by_path": byPath,
			"avg_latency_ms":   avgLatency,
		})
	}
}
