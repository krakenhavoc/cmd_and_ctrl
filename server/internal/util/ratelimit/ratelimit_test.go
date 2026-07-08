package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAllowBurstThenRefill(t *testing.T) {
	l := New(1, 3)
	// Freeze time so refill is deterministic.
	now := time.Unix(1_000_000, 0)
	l.now = func() time.Time { return now }

	for i := 0; i < 3; i++ {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("burst %d: unexpectedly throttled", i)
		}
	}
	if l.Allow("1.2.3.4") {
		t.Error("4th immediate call should be throttled")
	}
	// Different key has its own bucket.
	if !l.Allow("5.6.7.8") {
		t.Error("different key should have full bucket")
	}
	// Advance 2s → 2 tokens back.
	now = now.Add(2 * time.Second)
	if !l.Allow("1.2.3.4") {
		t.Error("after refill: expected allow")
	}
	if !l.Allow("1.2.3.4") {
		t.Error("after refill: expected second allow")
	}
	if l.Allow("1.2.3.4") {
		t.Error("after burst drain: expected throttle")
	}
}

func TestMiddleware429(t *testing.T) {
	l := New(0.01, 1) // almost no refill, burst 1
	h := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request allowed.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "9.9.9.9:1111"
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("first: got %d, want 200", rec.Code)
	}

	// Second immediate request throttled.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "9.9.9.9:2222"
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("second: got %d, want 429", rec.Code)
	}
}

// TestKeyForTrustForwarded covers the CMDCTRL_TRUST_FORWARDED opt-in:
// off → RemoteAddr host; on → the proxy-appended (last) X-Forwarded-For
// hop, never the spoofable client-supplied prefix; on with no header →
// RemoteAddr fallback.
func TestKeyForTrustForwarded(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "10.0.0.1:4242"
	// The client sent its own "X-Forwarded-For: spoofed.example"; the
	// trusted proxy appended the real client IP after it.
	req.Header.Set("X-Forwarded-For", "spoofed.example, 203.0.113.7")

	// Default (unset / falsy): header ignored, proxy address keys.
	t.Setenv("CMDCTRL_TRUST_FORWARDED", "")
	if got := keyFor(req); got != "10.0.0.1" {
		t.Errorf("untrusted keyFor: got %q, want %q", got, "10.0.0.1")
	}
	t.Setenv("CMDCTRL_TRUST_FORWARDED", "0")
	if got := keyFor(req); got != "10.0.0.1" {
		t.Errorf("falsy keyFor: got %q, want %q", got, "10.0.0.1")
	}

	// Trusted: the last hop — the entry the proxy itself appended —
	// is the client; the attacker-controlled first hop is ignored.
	t.Setenv("CMDCTRL_TRUST_FORWARDED", "1")
	if got := keyFor(req); got != "203.0.113.7" {
		t.Errorf("trusted keyFor: got %q, want %q", got, "203.0.113.7")
	}

	// A single-entry header (honest client, one trusted proxy) keys
	// on that entry.
	honest := httptest.NewRequest(http.MethodPost, "/", nil)
	honest.RemoteAddr = "10.0.0.1:4242"
	honest.Header.Set("X-Forwarded-For", "198.51.100.9")
	if got := keyFor(honest); got != "198.51.100.9" {
		t.Errorf("trusted single-hop keyFor: got %q, want %q", got, "198.51.100.9")
	}

	// Trusted but header absent → RemoteAddr fallback.
	bare := httptest.NewRequest(http.MethodPost, "/", nil)
	bare.RemoteAddr = "10.0.0.2:9999"
	if got := keyFor(bare); got != "10.0.0.2" {
		t.Errorf("trusted no-header keyFor: got %q, want %q", got, "10.0.0.2")
	}
}

// TestSweepDropsStaleBuckets proves the on-access sweep reclaims
// buckets idle long enough to have refilled to a full burst, while
// keeping active ones (with their drained token state) intact.
func TestSweepDropsStaleBuckets(t *testing.T) {
	l := New(1, 3)
	now := time.Unix(1_000_000, 0)
	l.now = func() time.Time { return now }

	// Drain one key, lightly touch another.
	for i := 0; i < 3; i++ {
		l.Allow("stale.key")
	}
	l.Allow("fresh.key")
	if got := len(l.buckets); got != 2 {
		t.Fatalf("buckets before sweep: got %d, want 2", got)
	}

	// Advance past both the sweep interval and the full-refill
	// horizon (burst/rate = 3 s) for stale.key only: re-touch
	// fresh.key just before the sweeping call so it stays live.
	now = now.Add(sweepInterval - time.Second)
	l.Allow("fresh.key")
	now = now.Add(2 * time.Second) // total > sweepInterval since lastSweep=0 epoch
	l.Allow("trigger.key")         // this call runs the sweep

	l.mu.Lock()
	_, staleAlive := l.buckets["stale.key"]
	_, freshAlive := l.buckets["fresh.key"]
	l.mu.Unlock()
	if staleAlive {
		t.Error("stale bucket survived the sweep")
	}
	if !freshAlive {
		t.Error("recently-used bucket was swept")
	}
}
