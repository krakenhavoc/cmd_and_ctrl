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
