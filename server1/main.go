package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-ondemand/server1/config"
	"api-ondemand/server1/handlers"
	"api-ondemand/server1/metrics"
	"api-ondemand/server1/middleware"
)

func main() {
	port := config.GetEnv("PORT", config.DefaultPort)
	serverName := config.GetEnv("SERVER_NAME", config.DefaultServerName)
	startTime := time.Now()

	m := metrics.NewMetrics()

	mux := http.NewServeMux()
	mux.Handle("/", handlers.RootHandler(serverName, port))
	mux.Handle("/ping", handlers.PingHandler(serverName))
	mux.Handle("/info", handlers.InfoHandler(serverName, port, startTime))
	mux.Handle("/health", handlers.HealthHandler(serverName))
	mux.Handle("/echo", handlers.EchoHandler())
	mux.Handle("/metrics", handlers.MetricsHandler(m))

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      middleware.LoggingMiddleware(m, mux),
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
