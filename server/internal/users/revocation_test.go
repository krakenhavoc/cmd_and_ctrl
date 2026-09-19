package users

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

// countingStore wraps a WatermarkStore and counts calls, so a test can
// prove the request path never reaches the database.
type countingStore struct {
	WatermarkStore
	loads, writes atomic.Int32
	failWrites    bool
}

func (c *countingStore) SessionWatermarks(ctx context.Context) (map[uuid.UUID]time.Time, error) {
	c.loads.Add(1)
	return c.WatermarkStore.SessionWatermarks(ctx)
}

func (c *countingStore) InvalidateSessions(ctx context.Context, id uuid.UUID, at time.Time) (time.Time, error) {
	c.writes.Add(1)
	if c.failWrites {
		return time.Time{}, errors.New("disk full")
	}
	return c.WatermarkStore.InvalidateSessions(ctx, id, at)
}

func TestRevokeAllWritesTheWatermarkAndTheCache(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	u, err := s.UpsertFromDiscord(ctx, alice, "", "identify")
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	cs := &countingStore{WatermarkStore: s}
	r, err := NewRevocations(ctx, cs)
	if err != nil {
		t.Fatalf("NewRevocations: %v", err)
	}
	at := time.Date(2026, 9, 19, 12, 0, 0, 123_456_789, time.UTC)
	r.now = func() time.Time { return at }

	issued := at.Add(-time.Hour)
	if r.Revoked(u.ID, issued) {
		t.Fatal("revoked before anything was revoked")
	}

	stored, err := r.RevokeAll(ctx, u.ID)
	if err != nil {
		t.Fatalf("RevokeAll: %v", err)
	}
	if want := at.Truncate(time.Millisecond); !stored.Equal(want) {
		t.Errorf("stored watermark %v, want %v (milliseconds)", stored, want)
	}

	// The cache is invalidated by the write: the very next check sees
	// it, with no reload.
	if !r.Revoked(u.ID, issued) {
		t.Error("old session still valid after RevokeAll")
	}
	if r.Revoked(u.ID, stored.Add(time.Millisecond)) {
		t.Error("a session issued after the watermark is refused")
	}
	for i := 0; i < 100; i++ {
		r.Revoked(u.ID, issued)
		r.Revoked(uuid.New(), issued)
	}
	if n := cs.loads.Load(); n != 1 {
		t.Errorf("SessionWatermarks called %d times, want once at construction", n)
	}

	// And the database agrees.
	got, err := s.Get(ctx, u.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !got.SessionsInvalidBefore.Equal(stored) {
		t.Errorf("users row watermark %v, want %v", got.SessionsInvalidBefore, stored)
	}

	// A restart reads it back.
	r2, err := NewRevocations(ctx, s)
	if err != nil {
		t.Fatalf("NewRevocations after restart: %v", err)
	}
	if !r2.Revoked(u.ID, issued) {
		t.Error("revocation did not survive a restart")
	}
}

func TestWatermarkNeverMovesBack(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	u, _ := s.UpsertFromDiscord(ctx, alice, "", "identify")
	late := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	if _, err := s.InvalidateSessions(ctx, u.ID, late); err != nil {
		t.Fatalf("InvalidateSessions: %v", err)
	}
	got, err := s.InvalidateSessions(ctx, u.ID, late.Add(-time.Hour))
	if err != nil {
		t.Fatalf("InvalidateSessions earlier: %v", err)
	}
	if !got.Equal(late) {
		t.Errorf("watermark moved back to %v, want %v", got, late)
	}
}

func TestRevokeAllUnknownUser(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	r, err := NewRevocations(ctx, s)
	if err != nil {
		t.Fatalf("NewRevocations: %v", err)
	}
	ghost := uuid.New()
	if _, err := r.RevokeAll(ctx, ghost); !errors.Is(err, ErrNotFound) {
		t.Errorf("RevokeAll(unknown) err = %v, want ErrNotFound", err)
	}
	if r.Revoked(ghost, time.Now()) {
		t.Error("a failed revoke landed in the cache")
	}
}

func TestRevokeAllLeavesTheCacheAloneWhenTheWriteFails(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	u, _ := s.UpsertFromDiscord(ctx, alice, "", "identify")
	cs := &countingStore{WatermarkStore: s, failWrites: true}
	r, err := NewRevocations(ctx, cs)
	if err != nil {
		t.Fatalf("NewRevocations: %v", err)
	}
	if _, err := r.RevokeAll(ctx, u.ID); err == nil {
		t.Fatal("RevokeAll reported success over a failed write")
	}
	if r.Revoked(u.ID, time.Now().Add(-time.Hour)) {
		t.Error("cache says revoked though the database was never written")
	}
}

func TestSessionWatermarksSkipsUnrevokedUsers(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	if _, err := s.UpsertFromDiscord(ctx, alice, "", "identify"); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	m, err := s.SessionWatermarks(ctx)
	if err != nil {
		t.Fatalf("SessionWatermarks: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("watermarks = %v, want none for a user who never revoked", m)
	}
}
