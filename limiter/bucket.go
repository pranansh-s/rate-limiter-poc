package limiter

import (
	"math"
	"time"
)

type Bucket struct {
	tokens float64
	last   time.Time
}

func NewBucket(burst float64, now time.Time) *Bucket {
	return &Bucket{tokens: burst, last: now}
}

func (b *Bucket) Take(now time.Time, rate, burst float64) (bool, time.Duration, int) {
	elapsed := now.Sub(b.last).Seconds()
	b.tokens = math.Min(burst, b.tokens+elapsed*rate)
	b.last = now
	if b.tokens >= 1 {
		b.tokens--
		return true, 0, int(b.tokens)
	}
	wait := (1 - b.tokens) / rate
	retry := time.Duration(math.Ceil(wait*1000)) * time.Millisecond
	return false, retry, 0
}
