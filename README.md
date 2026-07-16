# rate-limiter-poc

A small HTTP rate limiter written in Go. It uses token buckets, supports
multiple layered limits, and can keep its counters in memory or in Redis.

## Quick start

```sh
go run .                          # in-memory counters, rules from config.json
go run . -store redis             # shared counters via Redis (REDIS_ADDR, default localhost:6379)
go run . -config other.json -addr :8081
docker compose up -d              # starts a local Redis
```

Try it out:

```sh
for i in $(seq 1 25); do curl -s -o /dev/null -w "%{http_code} " localhost:8080/ping; done
curl -i localhost:8080/ping
```

The first twenty or so requests return 200, after that you get 429 until the
bucket refills.

## Configuration

Rules live in `config.json`. A request has to pass every rule that applies
to it.

```json
{ "rules": [
  { "name": "ip-burst",  "scope": "ip",     "rate": 10,   "burst": 20 },
  { "name": "ip-hourly", "scope": "ip",     "rate": 0.28, "burst": 1000 },
  { "name": "key-burst", "scope": "key",    "rate": 50,   "burst": 100 },
  { "name": "global",    "scope": "global", "rate": 500,  "burst": 1000 }
] }
```

| scope | keyed by |
|-------|----------|
| `ip` | client IP address |
| `key` | the `X-API-Key` header, skipped when absent |
| `global` | one shared bucket for the whole service |

`rate` is tokens per second, `burst` is the bucket size. Two rules can share a
scope. That is how you get long term limits: `ip-burst` handles spikes while
`ip-hourly` (0.28/s is roughly 1000 per hour) catches the slow grind that a
burst rule never notices.

## How it works

Each client key gets a bucket of `burst` tokens. A request spends one token and
tokens flow back in at `rate` per second. Simple, and it has two nice
properties: clients can burst without being punished for it, and the long term
average stays capped.

Refill happens lazily. There are no timers and no goroutine per bucket. When a
request arrives, the limiter looks at how much time has passed and credits the
bucket accordingly. Idle clients cost nothing.

When a request is denied the response is `429 Too Many Requests` with a
`Retry-After` header. Every response, allowed or not, includes
`X-RateLimit-Limit`, `X-RateLimit-Remaining` and `X-RateLimit-Reset` so a well
behaved client can slow itself down before hitting the wall. If several rules
deny at once, the largest Retry-After wins.

## Benchmarks

Apache Bench, 50 concurrent connections, localhost, one rule at 200/s with
burst 200:

| store | throughput | allowed in 10s | p99 |
|-------|-----------|----------------|-----|
| memory | 42.5k req/s | 2192 (expected ~2200) | 3ms |
| redis | 32.7k req/s | 2194 (expected ~2200) | 4ms |

Enforcement is accurate to well under a percent. The memory store adds about 3%
to request time, and the single mutex never showed up as a bottleneck, so the
map stays unsharded on purpose.