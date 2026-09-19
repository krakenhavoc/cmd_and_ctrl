package discord

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestStateStoreRoundTrip(t *testing.T) {
	s := NewStateStore()
	game := uuid.New()
	invite := "secret-invite"

	state, challenge, err := s.Start(game, invite)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if state == "" || challenge == "" {
		t.Fatalf("empty state/challenge: %q %q", state, challenge)
	}

	// Consume returns the bound data and clears the entry.
	entry, err := s.Consume(state)
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}
	if entry.GameID != game {
		t.Errorf("GameID: got %v, want %v", entry.GameID, game)
	}
	if entry.InviteToken != invite {
		t.Errorf("InviteToken: got %q, want %q", entry.InviteToken, invite)
	}
	if entry.CodeVerifier == "" {
		t.Error("CodeVerifier should be non-empty")
	}
	// Second Consume is a miss — single-use semantics.
	if _, err := s.Consume(state); err != ErrStateNotFound {
		t.Errorf("replayed Consume: got %v, want ErrStateNotFound", err)
	}
}

func TestStateStoreRejectsForgedState(t *testing.T) {
	s := NewStateStore()
	if _, err := s.Consume("not-a-real-state"); err != ErrStateNotFound {
		t.Errorf("unknown state: got %v, want ErrStateNotFound", err)
	}
}

func TestStateStoreExpires(t *testing.T) {
	s := NewStateStore()
	// Shrink the TTL + stub the clock so we can step past expiry
	// without a real sleep.
	s.ttl = 100 * time.Millisecond
	now := time.Now()
	s.nowFn = func() time.Time { return now }

	state, _, err := s.Start(uuid.New(), "inv")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	// Advance the clock past the TTL — the next Start's gc drops it,
	// then the Consume call confirms the state is gone.
	now = now.Add(200 * time.Millisecond)
	_, _, _ = s.Start(uuid.New(), "other")

	if _, err := s.Consume(state); err != ErrStateNotFound {
		t.Errorf("expired Consume: got %v, want ErrStateNotFound", err)
	}
}

func TestStateStoreGCsOnConsume(t *testing.T) {
	// An idle store (no Starts) should still drop stale entries when
	// Consume is called — the gc fires opportunistically on both
	// paths so a never-again-called Start doesn't leak.
	s := NewStateStore()
	s.ttl = 50 * time.Millisecond
	now := time.Now()
	s.nowFn = func() time.Time { return now }

	state1, _, _ := s.Start(uuid.New(), "a")
	state2, _, _ := s.Start(uuid.New(), "b")

	now = now.Add(100 * time.Millisecond)

	// Consume of the second state should succeed even though both
	// are past TTL — wait, the gc drops both before the lookup. So
	// Consume returns ErrStateNotFound. That's the expected
	// semantics: expired states don't resurrect themselves for the
	// consumer that happened to hit the GC cycle first.
	if _, err := s.Consume(state2); err != ErrStateNotFound {
		t.Errorf("post-TTL Consume: got %v, want ErrStateNotFound", err)
	}
	// And the other stale entry is gone from the map.
	if _, err := s.Consume(state1); err != ErrStateNotFound {
		t.Errorf("post-gc map miss: got %v, want ErrStateNotFound", err)
	}
}

func TestStateStoreLinkRoundTrip(t *testing.T) {
	s := NewStateStore()
	game, player := uuid.New(), uuid.New()
	state, _, err := s.StartLink(game, player)
	if err != nil {
		t.Fatalf("StartLink: %v", err)
	}
	e, err := s.Consume(state)
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}
	if !e.Link() || e.Unbound() {
		t.Errorf("link entry: Link()=%v Unbound()=%v, want true/false", e.Link(), e.Unbound())
	}
	if e.GameID != game || e.LinkPlayerID != player || e.InviteToken != "" {
		t.Errorf("entry = %+v", e)
	}
	if e.CodeVerifier == "" {
		t.Error("link entry has no PKCE verifier")
	}
}

func TestStateStoreLinkNeedsASeat(t *testing.T) {
	s := NewStateStore()
	if _, _, err := s.StartLink(uuid.New(), uuid.Nil); err == nil {
		t.Error("StartLink with no player succeeded")
	}
	if _, _, err := s.StartLink(uuid.Nil, uuid.New()); err == nil {
		t.Error("StartLink with no game succeeded")
	}
	// The invite and unbound shapes are not links.
	st, _, _ := s.Start(uuid.New(), "tok")
	if e, _ := s.Consume(st); e.Link() {
		t.Error("an invite round-trip reports Link()")
	}
}
