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
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/envflag"
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
// key is derived from r.RemoteAddr (host portion only); deployments
// sitting behind a reverse proxy should set CMDCTRL_TRUST_FORWARDED
// to a truthy value so the key derives from the proxy-appended hop of
// X-Forwarded-For instead (RemoteAddr would collapse every caller
// into the proxy's single bucket).
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
// bucket. With CMDCTRL_TRUST_FORWARDED set (truthy), the LAST hop of
// X-Forwarded-For wins — that's the one entry the trusted reverse
// proxy appended itself. Everything left of it arrived in the
// client's own header, so keying on the first hop would let a caller
// mint a fresh bucket per request with a random X-Forwarded-For and
// bypass the limiter entirely. Falls back to RemoteAddr when the
// header is absent. The env var is read per request (matching the
// other CMDCTRL runtime knobs) so flipping it doesn't need a restart.
func keyFor(r *http.Request) string {
	if envflag.Truthy(os.Getenv("CMDCTRL_TRUST_FORWARDED")) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			last := xff
			if i := strings.LastIndexByte(xff, ','); i >= 0 {
				last = xff[i+1:]
			}
			if last = strings.TrimSpace(last); last != "" {
				return last
			}
		}
	}
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
