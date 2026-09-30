package game

import "testing"

// delayed_turn_of_test.go — DelayedTrigger.TurnOf, "at the beginning of
// THAT PLAYER's next end step" (The Eternal Wanderer's +1, #1538): a
// matching step on anyone else's turn leaves the trigger queued, and
// the delayed ability stays controlled by the player whose effect made
// it (CR 603.7d).

func TestDelayedTriggerTurnOfWaitsForTheNamedPlayersTurn(t *testing.T) {
	g := newActiveGame(t)
	if len(g.Seats) < 2 {
		t.Skip("needs two seats")
	}
	// Seat 0's effect, for seat 1's next end step.
	advanceTo(t, g, StepPrecombatMain)
	fired := 0
	var owner, controller = g.Seats[1].ID, g.Seats[0].ID
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: controller,
			Label:      "probe — that player's next end step",
			At:         StepEnd,
			TurnOf:     owner,
			Body: testBody(func(_ *Game, item *StackItem) error {
				fired++
				if item.Controller != controller {
					t.Errorf("fired item controller = %v, want the effect's controller %v", item.Controller, controller)
				}
				return nil
			}),
		})
	})

	// Seat 0's own end step matches the step but not the named turn.
	advanceToStepOfSeat(t, g, 0, StepEnd)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("queue=%d after the controller's end step, want the trigger still pending", len(g.DelayedTriggers))
	}
	if len(g.PendingTriggers)+len(g.StackMeta) != 0 {
		t.Fatal("the trigger reached the stack on the wrong player's turn")
	}

	advanceToStepOfSeat(t, g, 1, StepEnd)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("queue=%d on the named player's end step, want drained", len(g.DelayedTriggers))
	}
	settleStack(t, g)
	if fired != 1 {
		t.Errorf("effect ran %d times, want 1", fired)
	}
}

// Scheduled on the named player's own turn before the step, the very
// same turn's step counts: it is their next end step.
func TestDelayedTriggerTurnOfFiresThisTurnWhenItIsTheirs(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	fired := 0
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: g.Seats[0].ID,
			Label:      "probe",
			At:         StepEnd,
			TurnOf:     g.Seats[0].ID,
			Body: testBody(func(_ *Game, _ *StackItem) error {
				fired++
				return nil
			}),
		})
	})
	advanceToStepOfSeat(t, g, 0, StepEnd)
	settleStack(t, g)
	if fired != 1 {
		t.Errorf("effect ran %d times, want 1", fired)
	}
}

// Both gates are required when both are set.
func TestDelayedTriggerTurnOfComposesWithControllerTurnOnly(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller:         g.Seats[0].ID,
			Label:              "probe",
			At:                 StepEnd,
			ControllerTurnOnly: true,
			TurnOf:             g.Seats[1].ID,
			Body:               testBody(func(_ *Game, _ *StackItem) error { return nil }),
		})
	})
	// No turn satisfies "the controller's turn" and "seat 1's turn".
	advanceToStepOfSeat(t, g, 1, StepEnd)
	advanceToStepOfSeat(t, g, 0, StepEnd)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("queue=%d, want the impossible pair to stay queued", len(g.DelayedTriggers))
	}
}

// A player who left never has another turn (CR 800.4a): the trigger is
// dropped rather than queued forever.
func TestDelayedTriggerTurnOfIsDroppedWhenThePlayerLeaves(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: g.Seats[0].ID,
			Label:      "probe",
			At:         StepEnd,
			TurnOf:     g.Seats[1].ID,
			Body:       testBody(func(_ *Game, _ *StackItem) error { return nil }),
		})
	})
	g.WithWriteLock(func() { g.Seats[1].Eliminated = true })
	advanceToStepOfSeat(t, g, 0, StepEnd)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("queue=%d, want the departed player's trigger dropped", len(g.DelayedTriggers))
	}
	if len(g.PendingTriggers)+len(g.StackMeta) != 0 {
		t.Fatal("a departed player's trigger reached the stack")
	}
}

func TestCloneCarriesTurnOf(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: g.Seats[0].ID,
			Label:      "probe",
			At:         StepEnd,
			TurnOf:     g.Seats[1].ID,
			Body:       testBody(func(_ *Game, _ *StackItem) error { return nil }),
		})
	})
	snap := g.Clone()
	if len(snap.DelayedTriggers) != 1 || snap.DelayedTriggers[0].TurnOf != g.Seats[1].ID {
		t.Fatalf("clone dropped TurnOf: %#v", snap.DelayedTriggers)
	}
}
