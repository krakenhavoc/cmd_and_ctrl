// Package ratelimit is a minimal per-key token-bucket limiter used to
// blunt brute-force attempts against the admin-login and invite-join
// endpoints. Avoids pulling in golang.org/x/time/rate so the server's
// dependency footprint stays tight.
//
// Buckets that have sat idle long enough to refill to a full burst
// are indistinguishable from fresh ones, so Allow opportunistically
// sweeps them every sweepInterval — the map stays bounded by the set
// of recently-active keys instead of growing for the process
// lifetime.
package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/clientip"
)

// sweepInterval bounds how often Allow walks the whole bucket map
// looking for stale entries. Five minutes keeps the amortised cost
// negligible while still reclaiming abandoned keys promptly at
// personal-server scale.
const sweepInterval = 5 * time.Minute

// Limiter is a per-key token bucket.
type Limiter struct {
	// Rate is tokens added per second.
	Rate float64
	// Burst is the bucket capacity.
	Burst float64

	mu        sync.Mutex
	buckets   map[string]*bucket
	lastSweep time.Time
	now       func() time.Time // injected for tests
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
	l.sweepLocked(now)
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

// sweepLocked drops buckets that have been idle long enough to have
// refilled to a full burst — behaviourally identical to a fresh
// bucket, so deleting them changes nothing for the key while keeping
// the map bounded. Runs at most once per sweepInterval; caller must
// hold l.mu. A non-positive Rate never refills, so sweeping would
// reset a permanently-drained bucket — skip entirely.
func (l *Limiter) sweepLocked(now time.Time) {
	if l.Rate <= 0 || now.Sub(l.lastSweep) < sweepInterval {
		return
	}
	l.lastSweep = now
	for k, b := range l.buckets {
		if now.Sub(b.last).Seconds()*l.Rate >= l.Burst {
			delete(l.buckets, k)
		}
	}
}

// Middleware wraps next and enforces the limiter keyed by client IP.
// A throttled request returns 429 with a short JSON error body. The
// key is clientip.FromRequest: r.RemoteAddr's host, or, with
// CMDCTRL_TRUST_FORWARDED set, the address Cloudflare saw (#3334).
// Without that flag behind a proxy, RemoteAddr would collapse every
// caller into the proxy's single bucket.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.AllowRequest(r) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"too many requests"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// AllowRequest is Allow keyed the way Middleware keys a request: by
// client IP. It lets a handler spend a token from a bucket that guards
// another route, only on the path that needs it (POST /me/decks spends
// one of POST /deck-coverage's tokens only when it must fetch a deck,
// ADR 0112 §3 item 4).
func (l *Limiter) AllowRequest(r *http.Request) bool {
	return l.Allow(clientip.FromRequest(r))
}

func minf(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
