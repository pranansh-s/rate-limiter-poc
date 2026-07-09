package main

import (
	"log"
	"net/http"

	"github.com/prananshsingh/rate-limiter-poc/limiter"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong\n"))
	})

	store := limiter.NewMemoryStore()
	handler := limiter.Limit(store, 10, 20, mux)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
