package limiter

import (
	"math"
	"net"
	"net/http"
	"strconv"
)

func Limit(store *MemoryStore, rate, burst float64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := store.Allow("ip:"+clientIP(r), rate, burst)

		h := w.Header()
		h.Set("X-RateLimit-Limit", strconv.Itoa(res.Limit))
		h.Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
		h.Set("X-RateLimit-Reset", strconv.FormatInt(res.Reset.Unix(), 10))

		if !res.OK {
			h.Set("Retry-After", strconv.Itoa(int(math.Ceil(res.RetryAfter.Seconds()))))
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
