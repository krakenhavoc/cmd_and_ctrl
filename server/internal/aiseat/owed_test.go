package aiseat

import (
	"testing"
	"time"
)

func TestRunnerOwedFor(t *testing.T) {
	r := &Runner{done: make(chan struct{})}
	now := time.Now()
	if got := r.OwedFor(now); got != 0 {
		t.Fatalf("a fresh runner owes %v, want 0", got)
	}
	r.markOwed()
	first := r.owedSince.Load()
	r.markOwed() // a second mark keeps the earlier time
	if r.owedSince.Load() != first {
		t.Fatal("markOwed moved the start of an existing debt")
	}
	if got := r.OwedFor(now.Add(3 * time.Minute)); got < 2*time.Minute+59*time.Second {
		t.Fatalf("owed %v after 3 minutes, want about 3m", got)
	}
	r.clearOwed()
	if got := r.OwedFor(now.Add(time.Hour)); got != 0 {
		t.Fatalf("owes %v after clearOwed, want 0", got)
	}
	r.markOwed()
	close(r.done)
	if got := r.OwedFor(now.Add(time.Hour)); got != 0 {
		t.Fatalf("an exited runner owes %v, want 0", got)
	}
}

func TestManagerMaxOwedIsZeroWithNoRunners(t *testing.T) {
	if got := NewManager(nil, nil).MaxOwed(); got != 0 {
		t.Fatalf("MaxOwed = %v, want 0", got)
	}
}
