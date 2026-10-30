package game

import (
	"testing"

	"github.com/google/uuid"
)

// turn_tally_became_tapped_test.go pins #4576: the per-object
// became-tapped tally (TurnTally.BecameTapped) that Captain America,
// Living Legend's "if it's the first time that creature has become
// tapped this turn" reads, and #4534: the cause a teamwork payment
// stamps on its tap.

func becameTapped(g *Game, id uuid.UUID) int {
	var n int
	g.WithWriteLock(func() { n = g.TimesBecameTappedThisTurn(id) })
	return n
}

func setTapped(t *testing.T, g *Game, id uuid.UUID, tapped bool) {
	t.Helper()
	if err := g.TapCard(id, tapped); err != nil {
		t.Fatalf("TapCard(%v): %v", tapped, err)
	}
}

// TestBecameTappedCountsOnlyRealTaps: a tap counts when the permanent
// actually became tapped (CR 701.26a), and tapping a tapped permanent,
// or untapping one, counts nothing.
func TestBecameTappedCountsOnlyRealTaps(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)

	if got := becameTapped(g, id); got != 0 {
		t.Fatalf("before any tap: %d, want 0", got)
	}
	setTapped(t, g, id, true)
	if got := becameTapped(g, id); got != 1 {
		t.Fatalf("after the first tap: %d, want 1", got)
	}
	setTapped(t, g, id, true) // already tapped: nothing happens
	if got := becameTapped(g, id); got != 1 {
		t.Errorf("tapping a tapped permanent counted: %d, want 1", got)
	}
	setTapped(t, g, id, false)
	if got := becameTapped(g, id); got != 1 {
		t.Errorf("an untap counted as a tap: %d, want 1", got)
	}
	setTapped(t, g, id, true)
	if got := becameTapped(g, id); got != 2 {
		t.Errorf("after the second real tap: %d, want 2", got)
	}
}

// TestBecameTappedIsPerObject: another permanent's taps are its own,
// and a permanent that left and came back is a new object with no taps
// of its own (CR 400.7).
func TestBecameTappedIsPerObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	a := pushScopedTestCreature(g, me, 2, 2)
	b := pushScopedTestCreature(g, me, 2, 2)

	setTapped(t, g, a, true)
	if got := becameTapped(g, b); got != 0 {
		t.Errorf("a's tap counted for b: %d", got)
	}
	bounceAndReturn(t, g, a, me)
	if got := becameTapped(g, a); got != 0 {
		t.Errorf("after the bounce: %d, want 0 — the permanent that came back is a new object", got)
	}
	setTapped(t, g, a, true)
	if got := becameTapped(g, a); got != 1 {
		t.Errorf("the new object's first tap: %d, want 1", got)
	}
}

// TestBecameTappedIsFlushedWithTheTurn: the tally dies at the turn
// boundary with the rest.
func TestBecameTappedIsFlushedWithTheTurn(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	setTapped(t, g, id, true)
	g.WithWriteLock(func() { g.resetTurnTallyLocked() })
	if got := becameTapped(g, id); got != 0 {
		t.Errorf("after the turn boundary: %d, want 0", got)
	}
}

// TestBecameTappedSurvivesCloneAndSnapshot: the tally is part of the
// turn's record, so undo (Clone / RestoreFrom) and a snapshot restore
// both keep it, and the clone does not share its map with the live game.
func TestBecameTappedSurvivesCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	setTapped(t, g, id, true)

	snap := g.Clone()
	setTapped(t, g, id, false)
	setTapped(t, g, id, true)
	if got := becameTapped(snap, id); got != 1 {
		t.Errorf("the clone shares its tally with the live game: %d, want 1", got)
	}
	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if got := becameTapped(g, id); got != 1 {
		t.Errorf("undo did not bring the tally back: %d, want 1", got)
	}

	_, restored := roundTrip(t, g)
	if got := becameTapped(restored, id); got != 1 {
		t.Errorf("a snapshot restore lost the tally: %d, want 1", got)
	}
	setTapped(t, restored, id, false)
	setTapped(t, restored, id, true)
	if got := becameTapped(restored, id); got != 2 {
		t.Errorf("the restored tally does not keep counting: %d, want 2", got)
	}
}

// TestBecameTappedIsReadFromAnAttackOnlyWhenTheAttackerTapped: the
// engine taps an attacker without an EventTapCard, so the declaration
// is where the tap is counted, and a vigilance attacker, which does not
// tap (CR 702.20), counts nothing.
func TestBecameTappedIsReadFromAnAttackOnlyWhenTheAttackerTapped(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	plain := pushScopedTestCreature(g, me, 2, 2)
	vigilant := pushScopedTestCreature(g, me, 2, 2)
	g.WithWriteLock(func() {
		g.findCardByIDLocked(plain).Tapped = true
		// Vigilance: the creature stays untapped through the declaration.
	})
	emit(g,
		Event{Kind: EventAttack, Actor: me, CardID: plain, Target: opp},
		Event{Kind: EventAttack, Actor: me, CardID: vigilant, Target: opp},
	)
	if got := becameTapped(g, plain); got != 1 {
		t.Errorf("a tapped attacker: %d, want 1", got)
	}
	if got := becameTapped(g, vigilant); got != 0 {
		t.Errorf("an attacker that did not tap: %d, want 0", got)
	}
}

// TestTeamworkPaymentStampsItsTapCause: the tap that pays a teamwork
// cost carries Event.TapCause, and no other tap does.
func TestTeamworkPaymentStampsItsTapCause(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	paid := pushScopedTestCreature(g, me, 2, 2)
	other := pushScopedTestCreature(g, me, 2, 2)

	g.WithWriteLock(func() { g.payTeamworkLocked(me, []uuid.UUID{paid}) })
	setTapped(t, g, other, true)

	causes := map[uuid.UUID]string{}
	for _, ev := range g.Events {
		if ev.Kind == EventTapCard {
			causes[ev.CardID] = ev.TapCause
		}
	}
	if causes[paid] != TapCauseTeamwork {
		t.Errorf("the teamwork tap's cause = %q, want %q", causes[paid], TapCauseTeamwork)
	}
	if c, ok := causes[other]; !ok || c != "" {
		t.Errorf("an ordinary tap's cause = %q (seen: %v), want empty", c, ok)
	}
}
