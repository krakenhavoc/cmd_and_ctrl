package game

import (
	"testing"

	"github.com/google/uuid"
)

// #2108 (ADR 0113's amendment of 2026-10-08): a granted change to a
// maximum hand size, stored on the player with a Duration.

// A granted change is read through its Duration, not only swept by it.
// Advance the turn counter WITHOUT running the sweep and the
// until-end-of-turn grant must already be over.
func TestGrantedHandSizeIsDurationCheckedByTheReader(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.GrantHandSizeForEffect(me.ID, HandSizeNoMaximum, 0, "t", uuid.Nil, g.UntilYourNextTurnDuration(me.ID)); err != nil {
			t.Fatal(err)
		}
	})
	if got := maxHandOf(g, me); got != NoMaxHandSize {
		t.Fatalf("before the next turn: %d, want no maximum", got)
	}
	g.WithWriteLock(func() { me.TurnsBegun++ })
	if got := maxHandOf(g, me); got != DefaultMaxHandSize {
		t.Errorf("after the turn ended, before any sweep: %d, want 7", got)
	}
}

// Several grants live at once, each at its own timestamp, and a clone
// (undo) owns its own list.
func TestGrantedHandSizesStackAndCloneIsolates(t *testing.T) {
	n := int64(0)
	restore := SetClockForTest(func() int64 { n += 10; return n })
	defer restore()
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		for i := 0; i < 2; i++ {
			if err := g.GrantHandSizeForEffect(me.ID, HandSizeModify, -3, "t", uuid.Nil, IndefiniteDuration()); err != nil {
				t.Fatal(err)
			}
		}
	})
	if got := maxHandOf(g, me); got != 1 {
		t.Errorf("two reductions of three: %d, want 1", got)
	}
	c := g.Clone()
	g.WithWriteLock(func() {
		if err := g.GrantHandSizeForEffect(me.ID, HandSizeSet, 5, "t", uuid.Nil, IndefiniteDuration()); err != nil {
			t.Fatal(err)
		}
	})
	if got := maxHandOf(g, me); got != 5 {
		t.Errorf("a later set: %d, want 5", got)
	}
	if got := maxHandOf(c, c.Seats[0]); got != 1 {
		t.Errorf("the clone saw the later grant: %d, want 1", got)
	}
}
