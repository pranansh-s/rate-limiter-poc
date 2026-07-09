package limiter

import (
	"sync"
	"time"
)

type Result struct {
	OK         bool
	RetryAfter time.Duration
	Remaining  int
	Limit      int
}

type entry struct {
	bucket     *Bucket
	staleAfter time.Duration
}

type MemoryStore struct {
	mu      sync.Mutex
	entries map[string]*entry
}

func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{entries: make(map[string]*entry)}
	go s.evictLoop(time.Minute)
	return s
}

func (s *MemoryStore) Allow(key string, rate, burst float64) Result {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[key]
	if !ok {
		e = &entry{
			bucket:     NewBucket(burst, now),
			staleAfter: time.Duration(burst / rate * float64(time.Second)),
		}
		s.entries[key] = e
	}
	allowed, retry, remaining := e.bucket.Take(now, rate, burst)
	return Result{OK: allowed, RetryAfter: retry, Remaining: remaining, Limit: int(burst)}
}

func (s *MemoryStore) evictLoop(interval time.Duration) {
	for now := range time.Tick(interval) {
		s.mu.Lock()
		for key, e := range s.entries {
			if now.Sub(e.bucket.last) > e.staleAfter {
				delete(s.entries, key)
			}
		}
		s.mu.Unlock()
	}
}
