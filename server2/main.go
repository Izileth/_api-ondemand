package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Server 2 - Port 8081] Received request: %s %s", r.Method, r.URL.Path)
	fmt.Fprintf(w, "Hello from Physical Server 2 on port 8081!")
}

func main() {
	addr := ":8081"
	http.HandleFunc("/", handler)
	log.Printf("Server 2 starting on %s...", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}
}
