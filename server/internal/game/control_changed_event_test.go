package game

import (
	"testing"

	"github.com/google/uuid"
)

// control_changed_event_test.go pins #930: one EventControlChanged,
// emitted from the one materialise step, for every way a layer-2
// control delta can happen.

// controlChangedEvents returns this game's control-change events, in
// order.
func controlChangedEvents(g *Game) []Event {
	var out []Event
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventControlChanged {
				out = append(out, ev)
			}
		}
	})
	return out
}

// TestGainControlEmitsControlChanged — the theft itself: one event,
// naming the permanent, the player who lost it, the player who gained
// it and the effect that took it.
func TestGainControlEmitsControlChanged(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushScopedTestCreature(g, opp.ID, 2, 2)
	spell := uuid.New()

	g.WithWriteLock(func() {
		g.GainControlForEffect(spell, victim, me.ID,
			g.UntilEndOfTurnDuration(), "test — steal")
	})
	g.ReadSnapshot(func() {})

	evs := controlChangedEvents(g)
	if len(evs) != 1 {
		t.Fatalf("control-change events: got %d, want 1", len(evs))
	}
	ev := evs[0]
	if ev.CardID != victim {
		t.Errorf("CardID = %s, want the stolen permanent %s", ev.CardID, victim)
	}
	if ev.Target != opp.ID {
		t.Errorf("Target (from) = %s, want the previous controller %s", ev.Target, opp.ID)
	}
	if ev.Actor != me.ID {
		t.Errorf("Actor (to) = %s, want the new controller %s", ev.Actor, me.ID)
	}
	if ev.Source != spell {
		t.Errorf("Source = %s, want the effect's source %s", ev.Source, spell)
	}
}

// TestControlRevertingAtExpiryEmitsControlChanged — Act of Treason's
// creature going home. The revert is a layer-2 delta like any other,
// so it emits, and it names no source because no effect applies any
// more: the permanent went back to its baseline.
func TestControlRevertingAtExpiryEmitsControlChanged(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushScopedTestCreature(g, opp.ID, 2, 2)

	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), victim, me.ID,
			g.UntilEndOfTurnDuration(), "test — steal")
	})
	g.ReadSnapshot(func() {})
	advancePastScopedCleanup(t, g)
	if got := controllerOfCard(t, g, victim); got != opp.ID {
		t.Fatalf("after cleanup: controller %s, want %s", got, opp.ID)
	}

	evs := controlChangedEvents(g)
	if len(evs) != 2 {
		t.Fatalf("control-change events: got %d, want 2 (the theft and the revert)", len(evs))
	}
	back := evs[1]
	if back.Target != me.ID || back.Actor != opp.ID {
		t.Errorf("revert: from %s to %s, want from %s to %s", back.Target, back.Actor, me.ID, opp.ID)
	}
	if back.Source != uuid.Nil {
		t.Errorf("revert: Source = %s, want nil — no effect applies any more", back.Source)
	}
}

// TestExchangeControlEmitsTwoEventsInOneBatch — CR 701.12 is one
// exchange, so the two deltas share the batch the pass ran in
// (CR 603.2c). A "whenever one or more" ability guarded by
// OncePerBatch has to see them as one occurrence.
func TestExchangeControlEmitsTwoEventsInOneBatch(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushScopedTestCreature(g, me.ID, 2, 2)
	theirs := pushScopedTestCreature(g, opp.ID, 3, 3)

	g.WithWriteLock(func() {
		if !g.ExchangeControlForEffect(uuid.New(), mine, theirs, "test — Switcheroo") {
			t.Fatal("ExchangeControlForEffect refused two battlefield permanents")
		}
	})
	g.ReadSnapshot(func() {})

	evs := controlChangedEvents(g)
	if len(evs) != 2 {
		t.Fatalf("control-change events: got %d, want 2", len(evs))
	}
	if evs[0].Batch != evs[1].Batch {
		t.Errorf("exchange spans batches %d and %d; one exchange is one occurrence (CR 603.2c)",
			evs[0].Batch, evs[1].Batch)
	}
	for _, ev := range evs {
		switch ev.CardID {
		case mine:
			if ev.Target != me.ID || ev.Actor != opp.ID {
				t.Errorf("mine: from %s to %s, want from %s to %s", ev.Target, ev.Actor, me.ID, opp.ID)
			}
		case theirs:
			if ev.Target != opp.ID || ev.Actor != me.ID {
				t.Errorf("theirs: from %s to %s, want from %s to %s", ev.Target, ev.Actor, opp.ID, me.ID)
			}
		default:
			t.Errorf("unexpected card %s in a two-permanent exchange", ev.CardID)
		}
	}
}

// TestControlEffectThatChangesNothingEmitsNoEvent — "gain control" of
// a permanent you already control is not a control change. The event
// is the DELTA, not the effect.
func TestControlEffectThatChangesNothingEmitsNoEvent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	mine := pushScopedTestCreature(g, me.ID, 2, 2)

	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), mine, me.ID,
			g.UntilEndOfTurnDuration(), "test — steal my own")
	})
	g.ReadSnapshot(func() {})
	// And again, to prove a second recompute over a standing effect
	// is silent too.
	g.BumpLayerVersionForTest()
	g.ReadSnapshot(func() {})

	if evs := controlChangedEvents(g); len(evs) != 0 {
		t.Errorf("control-change events: got %d, want 0 — the controller never changed", len(evs))
	}
}

// TestControlChangedSurvivesUndo — the event log is cloned with the
// game, so undoing back across a theft restores the board AND the
// history of how it got there, with nothing emitted twice.
func TestControlChangedSurvivesUndo(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushScopedTestCreature(g, opp.ID, 2, 2)

	before := g.Clone()
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), victim, me.ID,
			g.UntilEndOfTurnDuration(), "test — steal")
	})
	g.ReadSnapshot(func() {})
	if n := len(controlChangedEvents(g)); n != 1 {
		t.Fatalf("before undo: %d control-change events, want 1", n)
	}

	g.RestoreFrom(before)
	if got := controllerOfCard(t, g, victim); got != opp.ID {
		t.Errorf("after undo: controller %s, want %s", got, opp.ID)
	}
	if n := len(controlChangedEvents(g)); n != 0 {
		t.Errorf("after undo: %d control-change events, want 0", n)
	}
}
