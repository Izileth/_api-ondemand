package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Server 3 - Port 8082] Received request: %s %s", r.Method, r.URL.Path)
	fmt.Fprintf(w, "Hello from Physical Server 3 on port 8082!")
}

func main() {
	addr := ":8082"
	http.HandleFunc("/", handler)
	log.Printf("Server 3 starting on %s...", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}
}
