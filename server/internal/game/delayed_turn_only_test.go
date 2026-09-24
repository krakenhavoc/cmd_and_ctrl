package game

import "testing"

// delayed_turn_only_test.go — DelayedTrigger.ControllerTurnOnly, the
// "at the beginning of YOUR next main phase" gate Mana Drain needs:
// a matching step on another player's turn leaves the trigger
// queued; the controller's own turn fires it. Added with the
// card-coverage roadmap's batch 01 (#294).

// advanceToStepOfSeat walks the cursor until `seat` is active at
// `step`.
func advanceToStepOfSeat(t *testing.T, g *Game, seat int, step Step) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if g.Turn.Step == step && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward %s of seat %d: %v", step, seat, err)
		}
	}
	t.Fatalf("never reached %s of seat %d", step, seat)
}

func TestDelayedTriggerControllerTurnOnlySkipsOtherPlayersTurns(t *testing.T) {
	g := newActiveGame(t)
	if len(g.Seats) < 2 {
		t.Skip("needs two seats")
	}
	// Schedule from seat 0's turn, for seat 0, at the precombat
	// main — but only on seat 0's own turn.
	advanceTo(t, g, StepPrecombatMain)
	fired := 0
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller:         g.Seats[0].ID,
			Label:              "probe — your next main phase",
			At:                 StepPrecombatMain,
			ControllerTurnOnly: true,
			Body: testBody(func(_ *Game, _ *StackItem) error {
				fired++
				return nil
			}),
		})
	})

	// Seat 1's precombat main matches the step but not the owner.
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("queue=%d after an opponent's main phase, want the trigger still pending", len(g.DelayedTriggers))
	}
	if len(g.PendingTriggers)+len(g.StackMeta) != 0 {
		t.Fatal("the trigger reached the stack on an opponent's turn")
	}

	// Seat 0's next precombat main fires it.
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("queue=%d on the controller's main phase, want drained", len(g.DelayedTriggers))
	}
	settleStack(t, g)
	if fired != 1 {
		t.Errorf("effect ran %d times, want 1", fired)
	}
}

func TestDelayedTriggerWithoutTheGateFiresOnAnyTurn(t *testing.T) {
	g := newActiveGame(t)
	if len(g.Seats) < 2 {
		t.Skip("needs two seats")
	}
	advanceTo(t, g, StepPrecombatMain)
	fired := 0
	scheduleProbe(g, StepPrecombatMain, g.Seats[0].ID, &fired)

	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("queue=%d, want an ungated trigger to fire on the very next matching step", len(g.DelayedTriggers))
	}
	settleStack(t, g)
	if fired != 1 {
		t.Errorf("effect ran %d times, want 1", fired)
	}
}

func TestCloneCarriesControllerTurnOnly(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller:         g.Seats[0].ID,
			Label:              "probe",
			At:                 StepPrecombatMain,
			ControllerTurnOnly: true,
			Body:               testBody(func(_ *Game, _ *StackItem) error { return nil }),
		})
	})
	snap := g.Clone()
	if len(snap.DelayedTriggers) != 1 || !snap.DelayedTriggers[0].ControllerTurnOnly {
		t.Fatalf("clone dropped ControllerTurnOnly: %#v", snap.DelayedTriggers)
	}
}
