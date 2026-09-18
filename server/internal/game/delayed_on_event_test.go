package game

import (
	"testing"

	"github.com/google/uuid"
)

// delayed_on_event_test.go — #663: a delayed trigger whose condition
// is an EVENT rather than a step ("when you next cast an instant or
// sorcery spell this turn, copy that spell").
//
// The card half (Doublecast, Galvanic Iteration) is in the effects
// package; everything here is the mechanism: it fires on the first
// match and not on a non-match, it is gone once it fires, it is gone
// at end of turn unfired, two of them stack, and it goes through the
// harvester's dispatch rather than a second path of its own.

// scheduleOnCast queues an event-conditioned delayed trigger watching
// EventCast for `controller`, whose effect bumps *fired. `match` is
// an extra condition on the event; nil accepts every cast by the
// controller.
func scheduleOnCast(g *Game, controller uuid.UUID, fired *int, match func(Event) bool) uuid.UUID {
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: controller,
			Label:      "probe — when you next cast",
			On:         []EventKind{EventCast},
			AppliesTo: func(ev Event, dt *DelayedTrigger, _ *Game) bool {
				if ev.Actor != dt.Controller {
					return false
				}
				return match == nil || match(ev)
			},
			Effect: func(_ *Game, _ *StackItem) error {
				*fired++
				return nil
			},
		})
	})
	return id
}

// emitCast emits an EventCast by `actor` naming `cardID`, which is
// the event the trigger watches.
func emitCast(g *Game, actor, cardID uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventCast, Actor: actor, CardID: cardID})
	})
}

// TestEventDelayedTriggerFiresOnTheFirstMatchOnly is CR 603.7b: the
// trigger fires once and ceases to exist. A second matching cast in
// the same turn finds nothing owed.
func TestEventDelayedTriggerFiresOnTheFirstMatchOnly(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	fired := 0

	id := scheduleOnCast(g, me.ID, &fired, nil)
	if id == uuid.Nil {
		t.Fatal("ScheduleDelayedTriggerForEffect refused an event-conditioned trigger")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("queued %d delayed triggers, want 1", len(g.DelayedTriggers))
	}

	emitCast(g, me.ID, uuid.New())
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("the fired trigger is still queued (%d left)", len(g.DelayedTriggers))
	}
	settleStack(t, g)
	if fired != 1 {
		t.Fatalf("effect ran %d times after one cast, want 1", fired)
	}

	emitCast(g, me.ID, uuid.New())
	settleStack(t, g)
	if fired != 1 {
		t.Errorf("effect ran %d times after a second cast, want 1 — it fires once (CR 603.7b)", fired)
	}
}

// TestEventDelayedTriggerIgnoresANonMatchingEvent: the predicate is
// the whole condition. An opponent's cast is not "YOU next cast", and
// the trigger stays queued waiting for the one it names.
func TestEventDelayedTriggerIgnoresANonMatchingEvent(t *testing.T) {
	g := newActiveGame(t)
	me, opponent := g.Seats[0], g.Seats[1]
	fired := 0
	scheduleOnCast(g, me.ID, &fired, nil)

	emitCast(g, opponent.ID, uuid.New())
	settleStack(t, g)
	if fired != 0 {
		t.Fatalf("an opponent's cast fired the trigger (%d times)", fired)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("a non-matching event consumed the trigger (%d left)", len(g.DelayedTriggers))
	}

	// A watched kind with a predicate that says no is also not a
	// match — the kind is a pre-filter, not the condition.
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDrawCard, Actor: me.ID})
	})
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("an unwatched event kind consumed the trigger (%d left)", len(g.DelayedTriggers))
	}

	emitCast(g, me.ID, uuid.New())
	settleStack(t, g)
	if fired != 1 {
		t.Errorf("the trigger did not fire on the cast it names (%d)", fired)
	}
}

// TestTwoEventDelayedTriggersBothFireOnOneCast: two Doublecasts in a
// turn are two copies. Each trigger fires once; "once" is per
// trigger, not per event.
func TestTwoEventDelayedTriggersBothFireOnOneCast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	fired := 0
	scheduleOnCast(g, me.ID, &fired, nil)
	scheduleOnCast(g, me.ID, &fired, nil)

	emitCast(g, me.ID, uuid.New())
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("%d triggers left queued, want both fired", len(g.DelayedTriggers))
	}
	settleStack(t, g)
	if fired != 2 {
		t.Fatalf("effect ran %d times, want 2 — two copies", fired)
	}
}

// TestEventDelayedTriggerExpiresAtEndOfTurn is the duration half (CR
// 514.2): "this turn" ends at cleanup whether or not the cast ever
// came, and the next turn's cast finds nothing owed.
func TestEventDelayedTriggerExpiresAtEndOfTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	fired := 0
	scheduleOnCast(g, me.ID, &fired, nil)

	if d := g.DelayedTriggers[0].Duration; d == nil || d.Kind != UntilEndOfTurn {
		t.Fatalf("Duration = %+v, want an UntilEndOfTurn stamp", d)
	}

	advanceToStepOfSeat(t, g, 1, StepUpkeep)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("%d triggers survived the turn, want 0 (CR 514.2)", len(g.DelayedTriggers))
	}

	emitCast(g, me.ID, uuid.New())
	settleStack(t, g)
	if fired != 0 {
		t.Errorf("an expired trigger fired %d times", fired)
	}
}

// TestStepDelayedTriggerIsNotSweptAtEndOfTurn is the other side of
// that sweep: a step-conditioned trigger carries no duration, because
// "at the beginning of the NEXT end step" scheduled during an end
// step has to outlive the turn it was made in.
func TestStepDelayedTriggerIsNotSweptAtEndOfTurn(t *testing.T) {
	g := newActiveGame(t)
	fired := 0
	scheduleProbe(g, StepUpkeep, g.Seats[0].ID, &fired)

	if d := g.DelayedTriggers[0].Duration; d != nil {
		t.Fatalf("a step-conditioned trigger got a duration %+v, want none", d)
	}
	advanceToStepOfSeat(t, g, 1, StepUpkeep)
	settleStack(t, g)
	if fired != 1 {
		t.Errorf("the step trigger fired %d times across the turn boundary, want 1", fired)
	}
}

// TestEventDelayedTriggerGoesThroughTheHarvesterDispatch: an Optional
// one asks its controller, exactly as a harvested trigger does, and
// "no" drops it without effect. That the prompt exists at all is the
// proof that the fired trigger takes dispatchTriggerLocked rather
// than a second path.
func TestEventDelayedTriggerGoesThroughTheHarvesterDispatch(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	fired := 0
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: me.ID,
			Label:      "probe — you may copy that spell",
			On:         []EventKind{EventCast},
			Optional:   &TriggerOptionalPrompt{Question: "Copy it?"},
			AppliesTo: func(ev Event, dt *DelayedTrigger, _ *Game) bool {
				return ev.Actor == dt.Controller
			},
			Effect: func(*Game, *StackItem) error { fired++; return nil },
		})
	})

	emitCast(g, me.ID, uuid.New())
	choice := pendingChoiceOfKind(g, PendingChoiceTriggerPrompt)
	if choice == nil {
		t.Fatal("an Optional event-conditioned delayed trigger queued no prompt")
	}
	if choice.Chooser != me.ID {
		t.Errorf("prompt chooser = %s, want the trigger's controller %s", choice.Chooser, me.ID)
	}
	if err := g.ResolveTriggerPrompt(choice.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	settleStack(t, g)
	if fired != 0 {
		t.Errorf("a declined trigger ran its effect %d times", fired)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("a declined trigger stayed queued (%d) — it fired and was answered", len(g.DelayedTriggers))
	}
}

// TestEventDelayedTriggerCarriesTheTriggeringSpellAsPayload: "copy
// THAT spell" names the object the event named, and it reaches the
// Effect on the item rather than in a closure — which is what keeps
// the closure clone-safe (ADR 0026 §4).
func TestEventDelayedTriggerCarriesTheTriggeringSpellAsPayload(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	castID := uuid.New()
	var seen []TargetRef
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: me.ID,
			Label:      "probe — copy that spell",
			On:         []EventKind{EventCast},
			AppliesTo: func(ev Event, dt *DelayedTrigger, _ *Game) bool {
				return ev.Actor == dt.Controller
			},
			Effect: func(_ *Game, item *StackItem) error {
				seen = append([]TargetRef(nil), item.Payload...)
				return nil
			},
		})
	})

	emitCast(g, me.ID, castID)
	settleStack(t, g)

	if len(seen) != 1 || seen[0].Kind != TargetCard || seen[0].ID != castID {
		t.Fatalf("payload = %+v, want the cast spell %s", seen, castID)
	}
}

// TestEventDelayedTriggerSnapshotRoundTrip: the data half of the
// condition survives a snapshot so a restored game still knows WHAT
// was owed and until when; the closures do not, and the census says
// so rather than the snapshot pretending it is a restore point.
func TestEventDelayedTriggerSnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	fired := 0
	scheduleOnCast(g, me.ID, &fired, nil)

	snap := g.CaptureSnapshot()
	if snap.Continuations.DelayedTriggerEffects != 1 {
		t.Fatalf("census counted %d delayed-trigger effects, want 1",
			snap.Continuations.DelayedTriggerEffects)
	}
	if snap.Restorable() {
		t.Error("a snapshot holding a live delayed-trigger closure claims to be restorable")
	}

	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if len(restored.DelayedTriggers) != 1 {
		t.Fatalf("restored %d delayed triggers, want 1", len(restored.DelayedTriggers))
	}
	got := restored.DelayedTriggers[0]
	if len(got.On) != 1 || got.On[0] != EventCast {
		t.Errorf("restored On = %v, want [cast]", got.On)
	}
	if got.Duration == nil || got.Duration.Kind != UntilEndOfTurn {
		t.Errorf("restored Duration = %+v, want an UntilEndOfTurn stamp", got.Duration)
	}
	if got.Effect != nil || got.AppliesTo != nil {
		t.Error("a restored trigger came back with live closures — the census says they are dropped")
	}
}

// TestEventDelayedTriggerSurvivesClone: the queue is deep-copied, the
// closures shared, so a game rewound to before the cast still owes
// the trigger and replaying the cast fires it exactly once.
func TestEventDelayedTriggerSurvivesClone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	fired := 0
	scheduleOnCast(g, me.ID, &fired, nil)

	beforeCast := g.Clone()
	emitCast(g, me.ID, uuid.New())
	settleStack(t, g)
	if fired != 1 {
		t.Fatalf("effect ran %d times, want 1", fired)
	}

	fired = 0
	g.WithWriteLock(func() { g.RestoreFrom(beforeCast) })
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("the rewind left %d delayed triggers, want the one that was still owed", len(g.DelayedTriggers))
	}
	emitCast(g, me.ID, uuid.New())
	settleStack(t, g)
	if fired != 1 {
		t.Errorf("the replayed cast fired the trigger %d times, want 1", fired)
	}
}

// TestScheduleRefusesATriggerWithNeitherConditional keeps the
// malformed case dropped rather than queued, as it always was: a
// trigger with no step AND no event has nothing that could ever fire
// it and would sit in the queue forever.
func TestScheduleRefusesATriggerWithNeitherConditional(t *testing.T) {
	g := newActiveGame(t)
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: g.Seats[0].ID,
			Label:      "probe — no condition",
			Effect:     func(*Game, *StackItem) error { return nil },
		})
	})
	if id != uuid.Nil {
		t.Errorf("a trigger with no At and no On was queued (%s)", id)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d malformed triggers reached the queue", len(g.DelayedTriggers))
	}
}
