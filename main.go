package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/prananshsingh/rate-limiter-poc/limiter"
)

func main() {
	configPath := flag.String("config", "config.json", "path to rules config")
	flag.Parse()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong\n"))
	})

	store := limiter.NewMemoryStore()
	handler := limiter.Limit(store, cfg.Rules, mux)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
