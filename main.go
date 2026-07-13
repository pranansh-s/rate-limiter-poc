package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/prananshsingh/rate-limiter-poc/limiter"
)

func main() {
	configPath := flag.String("config", "config.json", "path to rules config")
	storeKind := flag.String("store", "memory", "memory or redis")
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong\n"))
	})

	var store limiter.Store
	switch *storeKind {
	case "memory":
		store = limiter.NewMemoryStore()
	case "redis":
		redisAddr := os.Getenv("REDIS_ADDR")
		if redisAddr == "" {
			redisAddr = "localhost:6379"
		}
		store = limiter.NewRedisStore(redisAddr)
	default:
		log.Fatalf("unknown store %q", *storeKind)
	}

	handler := limiter.Limit(store, cfg.Rules, mux)

	log.Println("listening on", *addr)
	log.Fatal(http.ListenAndServe(*addr, handler))
}
