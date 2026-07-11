package limiter

import (
	"math"
	"net"
	"net/http"
	"strconv"
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
		h := w.Header()

		for _, rule := range rules {
			res := store.Allow(rule.Name+"|ip:"+ip, rule.Rate, rule.Burst)

			h.Set("X-RateLimit-Limit", strconv.Itoa(res.Limit))
			h.Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
			h.Set("X-RateLimit-Reset", strconv.FormatInt(res.Reset.Unix(), 10))

			if !res.OK {
				h.Set("Retry-After", strconv.Itoa(int(math.Ceil(res.RetryAfter.Seconds()))))
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
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
