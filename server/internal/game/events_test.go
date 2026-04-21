package game

import (
	"testing"

	"github.com/google/uuid"
)

// TestEmitEventAppendsAndStampsSeq covers the baseline emit path:
// appending preserves order, Seq is monotonic starting at 1, and
// a freshly constructed Game starts with an empty log.
func TestEmitEventAppendsAndStampsSeq(t *testing.T) {
	g := newActiveGame(t)

	// Starting state: events from game setup are expected (mulligan
	// draw / initial untap). Snapshot the baseline so the test
	// asserts deltas, not absolutes — keeps it robust to future
	// mutations adding new emit points at Start().
	baseline := len(g.Events)
	baselineSeq := g.eventSeq

	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDealDamage, Amount: 3})
		g.EmitEvent(Event{Kind: EventChangeLife, Amount: -3})
	})

	if got := len(g.Events) - baseline; got != 2 {
		t.Fatalf("event log delta: got %d, want 2", got)
	}
	first := g.Events[baseline]
	second := g.Events[baseline+1]
	if first.Kind != EventDealDamage || first.Amount != 3 {
		t.Errorf("first event: %+v", first)
	}
	if second.Kind != EventChangeLife || second.Amount != -3 {
		t.Errorf("second event: %+v", second)
	}
	if first.Seq != baselineSeq+1 {
		t.Errorf("first Seq: got %d, want %d", first.Seq, baselineSeq+1)
	}
	if second.Seq != baselineSeq+2 {
		t.Errorf("second Seq: got %d, want %d", second.Seq, baselineSeq+2)
	}
}

// TestEmitEventMutationInstrumentation is the "a rules-visible
// mutation produces an event" pin. We exercise one representative
// from each instrumented path so a future regression (e.g. someone
// refactors drawCardLocked and loses the emit) gets caught here.
func TestEmitEventMutationInstrumentation(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	baseline := len(g.Events)

	// DrawCard → EventDrawCard.
	if err := g.DrawCard(p.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	if !containsEventKind(g.Events[baseline:], EventDrawCard) {
		t.Errorf("DrawCard did not emit EventDrawCard")
	}

	// ChangePlayerLife → EventChangeLife.
	cut := len(g.Events)
	if _, err := g.ChangePlayerLife(p.ID, -1); err != nil {
		t.Fatalf("ChangePlayerLife: %v", err)
	}
	if !containsEventKind(g.Events[cut:], EventChangeLife) {
		t.Errorf("ChangePlayerLife did not emit EventChangeLife")
	}

	// AddCounter on a real battlefield card → EventCounterPlaced.
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		Name:       "Test Creature",
		TypeLine:   "Creature",
		Power:      2,
		Toughness:  2,
		Owner:      p.ID,
		Controller: p.ID,
	})
	cut = len(g.Events)
	if err := g.AddCounter(cardID, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if !containsEventKind(g.Events[cut:], EventCounterPlaced) {
		t.Errorf("AddCounter did not emit EventCounterPlaced")
	}

	// TapCard → EventTapCard + EventUntapCard.
	cut = len(g.Events)
	if err := g.TapCard(cardID, true); err != nil {
		t.Fatalf("TapCard true: %v", err)
	}
	if !containsEventKind(g.Events[cut:], EventTapCard) {
		t.Errorf("TapCard(true) did not emit EventTapCard")
	}
	cut = len(g.Events)
	if err := g.TapCard(cardID, false); err != nil {
		t.Fatalf("TapCard false: %v", err)
	}
	if !containsEventKind(g.Events[cut:], EventUntapCard) {
		t.Errorf("TapCard(false) did not emit EventUntapCard")
	}

	// MarkDamage with positive delta → EventDealDamage.
	cut = len(g.Events)
	if err := g.MarkDamage(cardID, 1); err != nil {
		t.Fatalf("MarkDamage: %v", err)
	}
	if !containsEventKind(g.Events[cut:], EventDealDamage) {
		t.Errorf("MarkDamage did not emit EventDealDamage")
	}

	// ShuffleLibrary → EventSearchLibrary (Label=shuffle).
	cut = len(g.Events)
	if err := g.ShuffleLibrary(p.ID); err != nil {
		t.Fatalf("ShuffleLibrary: %v", err)
	}
	if !containsEventKind(g.Events[cut:], EventSearchLibrary) {
		t.Errorf("ShuffleLibrary did not emit EventSearchLibrary")
	}
}

// TestEmitEventCloneRoundTrip proves that Clone / RestoreFrom
// preserves the event log verbatim and continues the Seq counter
// at the snapshot point.
func TestEmitEventCloneRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDealDamage, Amount: 1})
		g.EmitEvent(Event{Kind: EventDealDamage, Amount: 2})
	})
	origCount := len(g.Events)
	origSeq := g.eventSeq

	clone := g.Clone()
	if len(clone.Events) != origCount {
		t.Errorf("clone event count: got %d, want %d", len(clone.Events), origCount)
	}
	if clone.eventSeq != origSeq {
		t.Errorf("clone eventSeq: got %d, want %d", clone.eventSeq, origSeq)
	}
	// Mutating the clone's log must not leak back into the original.
	clone.WithWriteLock(func() {
		clone.EmitEvent(Event{Kind: EventDealDamage, Amount: 99})
	})
	if len(g.Events) != origCount {
		t.Errorf("original event count mutated after clone emit")
	}
	if g.eventSeq != origSeq {
		t.Errorf("original eventSeq mutated after clone emit")
	}
	if clone.eventSeq != origSeq+1 {
		t.Errorf("clone eventSeq after emit: got %d, want %d", clone.eventSeq, origSeq+1)
	}
}

// TestEmitEventRestoreFromReplacesLog covers the undo-stack path:
// RestoreFrom must adopt src's event log AND eventSeq counter so
// a continued game doesn't reuse Seq values after a rewind.
func TestEmitEventRestoreFromReplacesLog(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDealDamage, Amount: 1})
	})
	snap := g.Clone()
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDealDamage, Amount: 2})
		g.EmitEvent(Event{Kind: EventDealDamage, Amount: 3})
	})
	if len(g.Events) != len(snap.Events)+2 {
		t.Fatalf("divergent event counts before restore")
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if len(g.Events) != len(snap.Events) {
		t.Errorf("RestoreFrom did not truncate event log")
	}
	if g.eventSeq != snap.eventSeq {
		t.Errorf("RestoreFrom did not restore eventSeq")
	}

	// Post-restore emits continue the sequence from the snapshot,
	// not from the truncated future.
	postSeq := g.eventSeq
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDealDamage, Amount: 9})
	})
	if g.Events[len(g.Events)-1].Seq != postSeq+1 {
		t.Errorf("post-restore emit Seq: got %d, want %d", g.Events[len(g.Events)-1].Seq, postSeq+1)
	}
}

func containsEventKind(events []Event, kind EventKind) bool {
	for _, ev := range events {
		if ev.Kind == kind {
			return true
		}
	}
	return false
}
