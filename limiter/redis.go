package limiter

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const takeScript = `
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local t = redis.call('HMGET', KEYS[1], 'tokens', 'ts')
local tokens = tonumber(t[1]) or burst
local ts = tonumber(t[2]) or now
tokens = math.min(burst, tokens + (now - ts) / 1000 * rate)
local ok = 0
if tokens >= 1 then
  tokens = tokens - 1
  ok = 1
end
redis.call('HSET', KEYS[1], 'tokens', tokens, 'ts', now)
redis.call('PEXPIRE', KEYS[1], math.ceil(burst / rate * 1000))
return {ok, tostring(tokens)}
`

type RedisStore struct {
	client *redis.Client
	script *redis.Script
}

func NewRedisStore(addr string) *RedisStore {
	return &RedisStore{
		client: redis.NewClient(&redis.Options{Addr: addr}),
		script: redis.NewScript(takeScript),
	}
}

func (s *RedisStore) Allow(ctx context.Context, key string, rate, burst float64) (Result, error) {
	now := time.Now()
	vals, err := s.script.Run(ctx, s.client, []string{key}, rate, burst, now.UnixMilli()).Slice()
	if err != nil {
		return Result{}, err
	}

	allowed := vals[0].(int64) == 1
	tokens, err := strconv.ParseFloat(vals[1].(string), 64)
	if err != nil {
		return Result{}, err
	}

	res := Result{
		OK:        allowed,
		Remaining: int(tokens),
		Limit:     int(burst),
		Reset:     now.Add(time.Duration((burst - tokens) / rate * float64(time.Second))),
	}
	if !allowed {
		res.RetryAfter = time.Duration(math.Ceil((1-tokens)/rate*1000)) * time.Millisecond
	}
	return res, nil
}
