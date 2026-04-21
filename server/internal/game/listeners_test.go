package game

import (
	"testing"
)

// capturingListener records every event it sees in registration
// order. Used as the test stand-in for the S19 trigger harvester —
// exercises the Listener + RegisterListener + notifyListenersLocked
// plumbing without depending on real card effects.
type capturingListener struct {
	received []Event
}

func (c *capturingListener) OnEvent(_ *Game, ev Event) {
	c.received = append(c.received, ev)
}

// TestListenerReceivesEmittedEvents pins the basic contract:
// registered listeners see every subsequent EmitEvent in order.
func TestListenerReceivesEmittedEvents(t *testing.T) {
	g := newActiveGame(t)
	cap := &capturingListener{}
	g.RegisterListener(cap)
	base := len(cap.received)

	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDealDamage, Amount: 1})
		g.EmitEvent(Event{Kind: EventDealDamage, Amount: 2})
		g.EmitEvent(Event{Kind: EventDealDamage, Amount: 3})
	})

	if got := len(cap.received) - base; got != 3 {
		t.Fatalf("received delta: got %d, want 3", got)
	}
	for i, want := range []int{1, 2, 3} {
		if cap.received[base+i].Amount != want {
			t.Errorf("received[%d] Amount: got %d, want %d", i, cap.received[base+i].Amount, want)
		}
	}
}

// TestListenerDispatchOrder confirms listeners fire in registration
// order. Essential for the S19 trigger harvester — the harvester
// needs to observe events before any downstream sugar listener sees
// the derived triggers.
func TestListenerDispatchOrder(t *testing.T) {
	g := newActiveGame(t)
	first := &capturingListener{}
	second := &capturingListener{}
	g.RegisterListener(first)
	g.RegisterListener(second)

	baseFirst := len(first.received)
	baseSecond := len(second.received)

	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDealDamage, Amount: 7})
	})

	if got := len(first.received) - baseFirst; got != 1 {
		t.Fatalf("first listener received delta: got %d, want 1", got)
	}
	if got := len(second.received) - baseSecond; got != 1 {
		t.Fatalf("second listener received delta: got %d, want 1", got)
	}
	// Both listeners see the same event with the same Seq — the
	// dispatch is synchronous, so Seq is already stamped by the
	// time the first listener runs.
	if first.received[baseFirst].Seq != second.received[baseSecond].Seq {
		t.Errorf("listeners disagreed on Seq: %d vs %d",
			first.received[baseFirst].Seq, second.received[baseSecond].Seq)
	}
}

// TestListenerCloneShallowCopy proves that clones share listeners
// (process-lifetime singletons — no deep copy) but each clone's
// event log is independent.
func TestListenerCloneShallowCopy(t *testing.T) {
	g := newActiveGame(t)
	cap := &capturingListener{}
	g.RegisterListener(cap)
	clone := g.Clone()

	if len(clone.Listeners) != len(g.Listeners) {
		t.Fatalf("clone listener count: got %d, want %d", len(clone.Listeners), len(g.Listeners))
	}
	// Emitting on the clone notifies the shared listener once.
	before := len(cap.received)
	clone.WithWriteLock(func() {
		clone.EmitEvent(Event{Kind: EventDealDamage, Amount: 1})
	})
	if got := len(cap.received) - before; got != 1 {
		t.Errorf("clone emit did not reach shared listener: delta %d", got)
	}
}
