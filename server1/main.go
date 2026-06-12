package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Server 1 - Port 8080] Received request: %s %s", r.Method, r.URL.Path)
	fmt.Fprintf(w, "Hello from Physical Server 1 on port 8080!")
}

func main() {
	addr := ":8080"
	http.HandleFunc("/", handler)
	log.Printf("Server 1 starting on %s...", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}
}
