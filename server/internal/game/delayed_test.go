package game

import (
	"testing"

	"github.com/google/uuid"
)

// delayed_test.go covers the S22 CR 603.7 delayed-trigger queue: it
// fires on entry to the step it names, it does NOT fire in the step
// it was created in ("the NEXT end step"), it goes through the stack
// rather than applying inline, and it survives Clone / RestoreFrom
// so undo neither loses a pending return nor resurrects a cancelled
// one.

// settleStack passes priority until nothing is on, or headed for,
// the stack. Mirrors the effects package's passPriorityAroundTable.
func settleStack(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if g.Stack.Size() == 0 && len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0 {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatalf("stack did not empty after 32 priority passes")
}

// scheduleProbe queues a delayed trigger firing at `at` whose effect
// bumps *fired. Returns the trigger's ID.
func scheduleProbe(g *Game, at Step, controller uuid.UUID, fired *int) uuid.UUID {
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: controller,
			Label:      "probe — delayed trigger",
			At:         at,
			Body: testBody(func(_ *Game, _ *StackItem) error {
				*fired++
				return nil
			}),
		})
	})
	return id
}

// TestDelayedTriggerFiresAtItsStep: a trigger scheduled before the
// end step is queued, untouched by intervening steps, and lands on
// the stack when the end step begins.
func TestDelayedTriggerFiresAtItsStep(t *testing.T) {
	g := newActiveGame(t)
	fired := 0
	id := scheduleProbe(g, StepEnd, g.Seats[0].ID, &fired)
	if id == uuid.Nil {
		t.Fatal("ScheduleDelayedTriggerForEffect returned a nil ID")
	}

	advanceTo(t, g, StepPrecombatMain)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed trigger drained before its step: queue=%d", len(g.DelayedTriggers))
	}
	if fired != 0 {
		t.Fatalf("effect ran early: fired=%d", fired)
	}

	advanceTo(t, g, StepEnd)
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("queue still holds %d triggers after the end step began", len(g.DelayedTriggers))
	}
	// It uses the stack: the item is waiting, not already applied.
	if fired != 0 {
		t.Errorf("effect applied without going through the stack: fired=%d", fired)
	}
	if len(g.StackMeta)+len(g.PendingTriggers) == 0 {
		t.Fatal("delayed trigger produced no stack item")
	}

	settleStack(t, g)
	if fired != 1 {
		t.Errorf("effect ran %d times, want 1", fired)
	}
}

// TestDelayedTriggerWaitsForTheNextOccurrence: "at the beginning of
// the NEXT end step" means a trigger created DURING an end step
// waits a whole turn. The queue is drained on step entry, so this
// falls out for free — pinned here because it is the timing the
// blink cards depend on.
func TestDelayedTriggerWaitsForTheNextOccurrence(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepEnd)

	fired := 0
	scheduleProbe(g, StepEnd, g.Seats[0].ID, &fired)

	// Still this same end step: nothing may fire.
	settleStack(t, g)
	if fired != 0 {
		t.Fatalf("trigger fired in the end step it was created in: fired=%d", fired)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("queue=%d, want the trigger still pending", len(g.DelayedTriggers))
	}

	// Walk off this end step, then on to the next one.
	advanceTo(t, g, StepUpkeep)
	advanceTo(t, g, StepEnd)
	settleStack(t, g)
	if fired != 1 {
		t.Errorf("trigger fired %d times at the next end step, want 1", fired)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("queue not drained: %d", len(g.DelayedTriggers))
	}
}

// TestDelayedTriggerAtUpkeep: the slot is not end-step-only. A
// trigger scheduled during an upkeep also waits for the NEXT one.
func TestDelayedTriggerAtUpkeep(t *testing.T) {
	g := newActiveGame(t)
	if g.Turn.Step != StepUpkeep {
		t.Fatalf("setup: cursor at %s, want upkeep", g.Turn.Step)
	}
	fired := 0
	scheduleProbe(g, StepUpkeep, g.Seats[1].ID, &fired)

	advanceTo(t, g, StepEnd)
	settleStack(t, g)
	if fired != 0 {
		t.Fatalf("upkeep trigger fired somewhere other than an upkeep: fired=%d", fired)
	}

	advanceTo(t, g, StepUpkeep)
	settleStack(t, g)
	if fired != 1 {
		t.Errorf("trigger fired %d times at the next upkeep, want 1", fired)
	}
}

// TestDelayedTriggerStampsPayloadOntoTheStackItem: the card IDs the
// trigger carries arrive as TargetCard refs on the item, which is
// how the effect reads them back (clone-safe) rather than closing
// over them.
func TestDelayedTriggerStampsPayloadOntoTheStackItem(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cardID := uuid.New()
	g.Exile.PushTop(Card{
		InstanceID: cardID,
		Name:       "Exiled Probe",
		TypeLine:   "Creature — Test",
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	var seen []uuid.UUID
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: owner.ID,
			Label:      "probe — payload",
			At:         StepEnd,
			Cards:      []uuid.UUID{cardID},
			Body: testBody(func(_ *Game, item *StackItem) error {
				for _, ref := range item.Targets {
					seen = append(seen, ref.ID)
				}
				return nil
			}),
		})
	})

	advanceTo(t, g, StepEnd)
	settleStack(t, g)
	if len(seen) != 1 || seen[0] != cardID {
		t.Errorf("effect saw payload %v, want [%v]", seen, cardID)
	}
}

// TestCloneAndRestorePreserveDelayedTriggers is the undo contract:
// a snapshot keeps the triggers pending at capture time, later
// scheduling does not leak into it, and RestoreFrom rewinds the
// queue — with the restored trigger still able to fire.
func TestCloneAndRestorePreserveDelayedTriggers(t *testing.T) {
	g := newActiveGame(t)
	fired := 0
	first := scheduleProbe(g, StepEnd, g.Seats[0].ID, &fired)

	snap := g.Clone()
	if len(snap.DelayedTriggers) != 1 {
		t.Fatalf("clone captured %d delayed triggers, want 1", len(snap.DelayedTriggers))
	}
	if snap.DelayedTriggers[0] == g.DelayedTriggers[0] {
		t.Error("clone aliased the original's DelayedTrigger pointer")
	}
	if snap.DelayedTriggers[0].ID != first {
		t.Errorf("clone trigger ID %v, want %v", snap.DelayedTriggers[0].ID, first)
	}

	// Post-snapshot scheduling must not reach the clone.
	scheduleProbe(g, StepEnd, g.Seats[0].ID, &fired)
	if len(g.DelayedTriggers) != 2 {
		t.Fatalf("live queue=%d, want 2", len(g.DelayedTriggers))
	}
	if len(snap.DelayedTriggers) != 1 {
		t.Errorf("post-snapshot schedule leaked into the clone: %d", len(snap.DelayedTriggers))
	}

	// Undo: rewind to the snapshot, then prove the restored trigger
	// still fires against the restored game.
	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].ID != first {
		t.Fatalf("RestoreFrom did not rewind the queue: %#v", g.DelayedTriggers)
	}

	advanceTo(t, g, StepEnd)
	settleStack(t, g)
	if fired != 1 {
		t.Errorf("restored trigger fired %d times, want 1", fired)
	}
}

// TestScheduleDelayedTriggerRejectsMalformed: a request with no
// effect or no step is dropped rather than queued, so a bad card
// file can never wedge the queue with something that never fires.
func TestScheduleDelayedTriggerRejectsMalformed(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		if id := g.ScheduleDelayedTriggerForEffect(DelayedTrigger{At: StepEnd}); id != uuid.Nil {
			t.Errorf("queued a trigger with no Effect: %v", id)
		}
		if id := g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Body: testBody(func(_ *Game, _ *StackItem) error { return nil }),
		}); id != uuid.Nil {
			t.Errorf("queued a trigger with no step: %v", id)
		}
	})
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("queue=%d, want empty", len(g.DelayedTriggers))
	}
}

// TestEndStepEmitsBeginEndStep pins the event the "at the beginning
// of your end step" triggers watch, and that it names the active
// player.
func TestEndStepEmitsBeginEndStep(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepEnd)
	active := g.Seats[g.Turn.ActiveSeat].ID
	found := false
	for _, ev := range g.Events {
		if ev.Kind == EventBeginEndStep && ev.Actor == active {
			found = true
		}
	}
	if !found {
		t.Error("entering the end step emitted no begin_end_step for the active player")
	}
}
