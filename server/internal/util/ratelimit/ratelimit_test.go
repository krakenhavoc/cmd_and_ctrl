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

// TestMiddlewareKeysOnTheClientBehindCloudflare is #3334: behind the
// hosts' proxies every request arrives from loopback, and two visitors
// must not share a bucket. The header chain itself is pinned in
// internal/util/clientip.
func TestMiddlewareKeysOnTheClientBehindCloudflare(t *testing.T) {
	t.Setenv("CMDCTRL_TRUST_FORWARDED", "1")
	l := New(0, 1)
	h := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	hit := func(cf string) int {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.RemoteAddr = "127.0.0.1:51234"
		req.Header.Set("X-Forwarded-For", "127.0.0.1")
		req.Header.Set("CF-Connecting-IP", cf)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	if got := hit("203.0.113.7"); got != http.StatusOK {
		t.Fatalf("first visitor: got %d, want 200", got)
	}
	if got := hit("198.51.100.9"); got != http.StatusOK {
		t.Errorf("second visitor shares the first one's bucket: got %d, want 200", got)
	}
	if got := hit("203.0.113.7"); got != http.StatusTooManyRequests {
		t.Errorf("first visitor again: got %d, want 429", got)
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
