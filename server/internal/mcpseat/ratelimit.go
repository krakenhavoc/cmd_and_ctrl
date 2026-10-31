package mcpseat

import (
	"context"
	"sync"
	"time"
)

// bucket is a token bucket: rate tokens per second, up to burst. The
// seat keeps its own limits (§8) because the WebSocket has none for
// actions; these are a courtesy to the table, not a security boundary.
type bucket struct {
	mu     sync.Mutex
	rate   float64
	burst  float64
	tokens float64
	last   time.Time
	now    func() time.Time
}

func newBucket(rate float64, burst int, now func() time.Time) *bucket {
	if now == nil {
		now = time.Now
	}
	return &bucket{rate: rate, burst: float64(burst), tokens: float64(burst), last: now(), now: now}
}

func (b *bucket) refill() {
	t := b.now()
	b.tokens += t.Sub(b.last).Seconds() * b.rate
	if b.tokens > b.burst {
		b.tokens = b.burst
	}
	b.last = t
}

// allow takes a token if one is there, and otherwise says how long until
// one will be.
func (b *bucket) allow() (bool, time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refill()
	if b.tokens >= 1 {
		b.tokens--
		return true, 0
	}
	need := (1 - b.tokens) / b.rate
	return false, time.Duration(need * float64(time.Second))
}

// wait blocks until a token is free or ctx ends.
func (b *bucket) wait(ctx context.Context) error {
	for {
		ok, d := b.allow()
		if ok {
			return nil
		}
		t := time.NewTimer(d)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
	}
}

// spanLimit allows n events in any span of per, the shape of the server's
// own limit on legal_moves_request (ws.legalMovesLimiter, n a second). The
// seat fetches several cut cards for one window on its own (#4737), and a
// request the server refused would leave a card cut; waiting its turn here
// costs at most the rest of the second. The server's refusal is still
// retried (requestMoves) for the clock skew between the two.
type spanLimit struct {
	mu   sync.Mutex
	per  time.Duration
	at   []time.Time
	next int
	now  func() time.Time
}

func newSpanLimit(n int, per time.Duration, now func() time.Time) *spanLimit {
	if now == nil {
		now = time.Now
	}
	return &spanLimit{per: per, at: make([]time.Time, n), now: now}
}

// wait blocks until one more event fits, records it, or ctx ends.
func (l *spanLimit) wait(ctx context.Context) error {
	for {
		l.mu.Lock()
		t := l.now()
		oldest := l.at[l.next]
		if oldest.IsZero() || t.Sub(oldest) >= l.per {
			l.at[l.next] = t
			l.next = (l.next + 1) % len(l.at)
			l.mu.Unlock()
			return nil
		}
		d := l.per - t.Sub(oldest)
		l.mu.Unlock()
		timer := time.NewTimer(d)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
