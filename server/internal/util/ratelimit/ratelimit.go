// Package ratelimit is a minimal per-key token-bucket limiter used to
// blunt brute-force attempts against the admin-login and invite-join
// endpoints. Avoids pulling in golang.org/x/time/rate so the server's
// dependency footprint stays tight.
//
// The implementation trades precision for simplicity: buckets are kept
// forever (keyed by IP) so a long-running server with many unique
// sources will grow without bound. For a personal-scale LAN server
// that's fine; swap in a sweeping implementation if the threat model
// ever changes.
package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// Limiter is a per-key token bucket.
type Limiter struct {
	// Rate is tokens added per second.
	Rate float64
	// Burst is the bucket capacity.
	Burst float64

	mu      sync.Mutex
	buckets map[string]*bucket
	now     func() time.Time // injected for tests
}

type bucket struct {
	tokens float64
	last   time.Time
}

// New returns a Limiter that refills `rate` tokens per second with a
// ceiling of `burst` tokens per key.
func New(rate, burst float64) *Limiter {
	return &Limiter{
		Rate:    rate,
		Burst:   burst,
		buckets: make(map[string]*bucket),
		now:     time.Now,
	}
}

// Allow attempts to consume a token for key. Returns true if one was
// available, false if the caller should be throttled.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.Burst, last: now}
		l.buckets[key] = b
	}
	// Refill by elapsed * rate, capped at burst.
	elapsed := now.Sub(b.last).Seconds()
	b.tokens = minf(l.Burst, b.tokens+elapsed*l.Rate)
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Middleware wraps next and enforces the limiter keyed by client IP.
// A throttled request returns 429 with a short JSON error body. The
// key is derived from r.RemoteAddr (host portion only); deployments
// sitting behind a reverse proxy should set CMDCTRL_TRUST_FORWARDED
// and extract X-Forwarded-For themselves.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow(keyFor(r)) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"too many requests"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// keyFor extracts a stable per-client key from the request. Host-only
// (strip the ephemeral port) so retries from the same source share a
// bucket.
func keyFor(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func minf(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
