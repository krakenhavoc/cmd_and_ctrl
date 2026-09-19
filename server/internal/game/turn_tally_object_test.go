package game

import (
	"testing"

	"github.com/google/uuid"
)

// turn_tally_object_test.go pins #936: the card-facing "only once each
// turn" tallies are per OBJECT (CR 400.7), and the CR 726 loop
// breaker that shares their (source, label) pair is still per CARD.

const objectTallyLabel = "Probe — once each turn"

// bounceAndReturn sends a battlefield permanent to its owner's hand
// and puts it straight back — the #630 shape (Teferi bounced by
// Venser and recast), which keeps the instance ID and so is exactly
// the case a map keyed by instance ID cannot tell apart. The blink
// helper is not used here on purpose: an exile return already mints a
// fresh instance ID (resetAsNewObjectLocked), so it never had the bug.
func bounceAndReturn(t *testing.T, g *Game, id, owner uuid.UUID) {
	t.Helper()
	hand := ZoneRef{Kind: ZoneHand, Owner: owner}
	field := ZoneRef{Kind: ZoneBattlefield}
	if err := g.MoveCardByID(field, hand, id); err != nil {
		t.Fatalf("bounce to hand: %v", err)
	}
	if err := g.MoveCardByID(hand, field, id); err != nil {
		t.Fatalf("return to the battlefield: %v", err)
	}
}

func triggeredThisTurn(g *Game, source uuid.UUID, label string) int {
	var n int
	g.WithWriteLock(func() { n = g.TriggeredThisTurn(source, label) })
	return n
}

func resolvedThisTurn(g *Game, source uuid.UUID, label string) int {
	var n int
	g.WithWriteLock(func() { n = g.ResolvedThisTurn(source, label) })
	return n
}

func emitTallyEvent(g *Game, kind EventKind, actor, source uuid.UUID, label string) {
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: kind, Actor: actor, Source: source, Label: label})
	})
}

// TestOnceEachTurnGatesAreFreshForTheReturningObject — CR 400.7. The
// permanent that comes back is a new object, so "if this is the first
// time this has happened this turn" is true again, for both the
// trigger tally and the resolution tally.
func TestOnceEachTurnGatesAreFreshForTheReturningObject(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind EventKind
		read func(*Game, uuid.UUID, string) int
	}{
		{"triggered", EventTrigger, triggeredThisTurn},
		{"resolved", EventResolve, resolvedThisTurn},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			id := pushScopedTestCreature(g, me.ID, 2, 2)

			emitTallyEvent(g, tc.kind, me.ID, id, objectTallyLabel)
			if got := tc.read(g, id, objectTallyLabel); got != 1 {
				t.Fatalf("before the bounce: %d, want 1", got)
			}

			bounceAndReturn(t, g, id, me.ID)

			if got := tc.read(g, id, objectTallyLabel); got != 0 {
				t.Errorf("after the bounce: %d, want 0 — the permanent that came back is a new "+
					"object (CR 400.7) and its once-each-turn clause may fire again", got)
			}
			emitTallyEvent(g, tc.kind, me.ID, id, objectTallyLabel)
			if got := tc.read(g, id, objectTallyLabel); got != 1 {
				t.Errorf("the new object's own count: %d, want 1", got)
			}
		})
	}
}

// TestTheOldObjectsTallyIsLeftBehindRatherThanDeleted — the fix is a
// key, not a cleanup. Nothing runs at the battlefield exit, the old
// entries simply stop being reachable, and the turn-boundary flush
// collects them. That is what keeps #935's one exit seam unchanged.
func TestTheOldObjectsTallyIsLeftBehindRatherThanDeleted(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushScopedTestCreature(g, me.ID, 2, 2)

	emitTallyEvent(g, EventTrigger, me.ID, id, objectTallyLabel)
	bounceAndReturn(t, g, id, me.ID)
	emitTallyEvent(g, EventTrigger, me.ID, id, objectTallyLabel)

	if n := len(g.TurnTally.Triggered); n != 2 {
		t.Errorf("Triggered holds %d entries, want 2 — one per object, neither deleted: %v",
			n, g.TurnTally.Triggered)
	}
	g.WithWriteLock(func() { g.resetTurnTallyLocked() })
	if n := len(g.TurnTally.Triggered); n != 0 {
		t.Errorf("the turn boundary left %d entries behind", n)
	}
}

// TestLoopBreakerStillTripsWhenTheSourceLeavesAndReturns is the
// regression the naive fix would have caused. Clearing the tally at
// the battlefield exit would have reset LoopRun on every iteration of
// a loop that blinks its own source, and the threshold would never be
// reached — the escape hatch created by exactly the loops the breaker
// exists for (ADR 0055).
func TestLoopBreakerStillTripsWhenTheSourceLeavesAndReturns(t *testing.T) {
	const threshold = 3
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	me := g.Seats[0]
	id := pushScopedTestCreature(g, me.ID, 2, 2)

	for i := 0; i < threshold; i++ {
		emitTallyEvent(g, EventResolve, me.ID, id, objectTallyLabel)
		// The card-facing gate is fresh every iteration...
		if got := resolvedThisTurn(g, id, objectTallyLabel); got != 1 {
			t.Fatalf("iteration %d: the object's own count is %d, want 1", i, got)
		}
		bounceAndReturn(t, g, id, me.ID)
	}

	// ...and the breaker's count is not.
	if got := g.TurnTally.LoopRun[TallyKey(id, objectTallyLabel)]; got != threshold {
		t.Errorf("loop run = %d after %d resolutions of one ability, want %d — LoopRun is keyed "+
			"per CARD so a blink loop cannot reset it", got, threshold, threshold)
	}
	if g.LoopNotice == nil {
		t.Fatal("the breaker never noticed a loop that leaves and re-enters the battlefield")
	}
	if g.LoopNotice.Source != id || g.LoopNotice.Label != objectTallyLabel {
		t.Errorf("notice names %s / %q, want %s / %q",
			g.LoopNotice.Source, g.LoopNotice.Label, id, objectTallyLabel)
	}
}

// TestObjectTallySurvivesUndo — the epoch rides the clone, so undoing
// back across the bounce puts the gate back where it was: shut.
func TestObjectTallySurvivesUndo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushScopedTestCreature(g, me.ID, 2, 2)
	emitTallyEvent(g, EventTrigger, me.ID, id, objectTallyLabel)

	before := g.Clone()
	bounceAndReturn(t, g, id, me.ID)
	if got := triggeredThisTurn(g, id, objectTallyLabel); got != 0 {
		t.Fatalf("after the bounce: %d, want 0", got)
	}

	g.RestoreFrom(before)
	if got := triggeredThisTurn(g, id, objectTallyLabel); got != 1 {
		t.Errorf("after undoing the bounce: %d, want 1 — the object that never left still "+
			"remembers its trigger", got)
	}
}

// TestObjectTallySurvivesASnapshotRoundTrip — the epoch is carried, so
// a game that is saved and restored mid-turn answers the same
// question the same way. A restore that dropped it would merge the
// returning object's counts with the ones before it.
func TestObjectTallySurvivesASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushScopedTestCreature(g, me.ID, 2, 2)
	bounceAndReturn(t, g, id, me.ID)
	emitTallyEvent(g, EventTrigger, me.ID, id, objectTallyLabel)
	if got := triggeredThisTurn(g, id, objectTallyLabel); got != 1 {
		t.Fatalf("before the round trip: %d, want 1", got)
	}

	_, restored := roundTrip(t, g)

	if got := triggeredThisTurn(restored, id, objectTallyLabel); got != 1 {
		t.Errorf("after the round trip: %d, want 1 — the tally key and the object's epoch must "+
			"still agree", got)
	}
	var epoch int
	restored.WithWriteLock(func() { epoch = restored.objectEpochLocked(id) })
	if epoch != 2 {
		t.Errorf("restored object epoch = %d, want 2 (one move out, one back)", epoch)
	}
}
