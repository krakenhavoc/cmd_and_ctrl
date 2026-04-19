package game

import (
	"testing"

	"github.com/google/uuid"
)

func TestSetPromiseRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]

	if err := g.SetPromise(p0.ID, p1.ID, 3); err != nil {
		t.Fatalf("SetPromise: %v", err)
	}
	if got := g.Promises[PromiseKey{From: p0.ID, To: p1.ID}]; got != 3 {
		t.Errorf("promise: got %d, want 3", got)
	}
	// Set semantics — overwrite, not increment.
	_ = g.SetPromise(p0.ID, p1.ID, 1)
	if got := g.Promises[PromiseKey{From: p0.ID, To: p1.ID}]; got != 1 {
		t.Errorf("promise overwrite: got %d, want 1", got)
	}
	// Negative clamps to 0.
	_ = g.SetPromise(p0.ID, p1.ID, -7)
	if got := g.Promises[PromiseKey{From: p0.ID, To: p1.ID}]; got != 0 {
		t.Errorf("promise negative clamp: got %d, want 0", got)
	}
	// Unknown player.
	if err := g.SetPromise(p0.ID, uuid.New(), 1); err != ErrPlayerNotFound {
		t.Errorf("promise unknown to: got %v, want ErrPlayerNotFound", err)
	}
}

func TestVoteHappyPath(t *testing.T) {
	g := newActiveGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]

	id, err := g.StartVote(p0.ID, "draw or discard?", []string{"draw", "discard"})
	if err != nil {
		t.Fatalf("StartVote: %v", err)
	}
	if g.Vote == nil || g.Vote.ID != id {
		t.Fatalf("StartVote did not stamp Vote on game")
	}
	if g.Vote.Topic != "draw or discard?" {
		t.Errorf("topic: got %q", g.Vote.Topic)
	}

	if err := g.CastVote(p0.ID, 0); err != nil {
		t.Fatalf("CastVote p0: %v", err)
	}
	if err := g.CastVote(p1.ID, 1); err != nil {
		t.Fatalf("CastVote p1: %v", err)
	}
	// Re-vote overwrites.
	if err := g.CastVote(p1.ID, 0); err != nil {
		t.Fatalf("CastVote p1 revote: %v", err)
	}
	if g.Vote.Ballots[p1.ID] != 0 {
		t.Errorf("revote: got %d, want 0", g.Vote.Ballots[p1.ID])
	}

	finished, err := g.EndVote()
	if err != nil {
		t.Fatalf("EndVote: %v", err)
	}
	if finished == nil || finished.ID != id {
		t.Fatal("EndVote did not return finished vote")
	}
	if g.Vote != nil {
		t.Error("EndVote did not clear Game.Vote")
	}
}

func TestVoteRejectsSecondStart(t *testing.T) {
	g := newActiveGame(t)
	p0 := g.Seats[0]
	if _, err := g.StartVote(p0.ID, "x", []string{"a", "b"}); err != nil {
		t.Fatalf("StartVote: %v", err)
	}
	if _, err := g.StartVote(p0.ID, "y", []string{"c", "d"}); err != ErrVoteAlreadyOpen {
		t.Errorf("second StartVote: got %v, want ErrVoteAlreadyOpen", err)
	}
}

func TestVoteRejectsTooFewOptions(t *testing.T) {
	g := newActiveGame(t)
	p0 := g.Seats[0]
	if _, err := g.StartVote(p0.ID, "x", []string{"only one"}); err != ErrInvalidParam {
		t.Errorf("one-option vote: got %v, want ErrInvalidParam", err)
	}
}

func TestCastVoteWithoutOpenVote(t *testing.T) {
	g := newActiveGame(t)
	p0 := g.Seats[0]
	if err := g.CastVote(p0.ID, 0); err != ErrNoVote {
		t.Errorf("cast without vote: got %v, want ErrNoVote", err)
	}
}

func TestCastVoteRejectsBadOption(t *testing.T) {
	g := newActiveGame(t)
	p0 := g.Seats[0]
	_, _ = g.StartVote(p0.ID, "x", []string{"a", "b"})
	if err := g.CastVote(p0.ID, 5); err != ErrInvalidOption {
		t.Errorf("bad option: got %v, want ErrInvalidOption", err)
	}
	if err := g.CastVote(p0.ID, -1); err != ErrInvalidOption {
		t.Errorf("negative option: got %v, want ErrInvalidOption", err)
	}
}

func TestEndVoteWithoutOpenVote(t *testing.T) {
	g := newActiveGame(t)
	if _, err := g.EndVote(); err != ErrNoVote {
		t.Errorf("end without vote: got %v, want ErrNoVote", err)
	}
}
