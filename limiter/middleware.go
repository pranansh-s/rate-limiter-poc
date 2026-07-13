package limiter

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Rule struct {
	Name  string  `json:"name"`
	Scope string  `json:"scope"`
	Rate  float64 `json:"rate"`
	Burst float64 `json:"burst"`
}

func Limit(store *MemoryStore, rules []Rule, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		apiKey := r.Header.Get("X-API-Key")

		var tightest *Result
		var retry time.Duration
		denied := false

		for _, rule := range rules {
			var scopeKey string
			switch rule.Scope {
			case "global":
				scopeKey = "global"
			case "ip":
				scopeKey = "ip:" + ip
			case "key":
				if apiKey == "" {
					continue
				}
				scopeKey = "key:" + apiKey
			}

			res := store.Allow(rule.Name+"|"+scopeKey, rule.Rate, rule.Burst)

			if !res.OK {
				denied = true
				if res.RetryAfter > retry {
					retry = res.RetryAfter
				}
			}
			if tightest == nil || res.Remaining < tightest.Remaining {
				tightest = &res
			}
		}

		h := w.Header()
		if tightest != nil {
			h.Set("X-RateLimit-Limit", strconv.Itoa(tightest.Limit))
			h.Set("X-RateLimit-Remaining", strconv.Itoa(tightest.Remaining))
			h.Set("X-RateLimit-Reset", strconv.FormatInt(tightest.Reset.Unix(), 10))
		}

		if denied {
			h.Set("Retry-After", strconv.Itoa(int(math.Ceil(retry.Seconds()))))
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
