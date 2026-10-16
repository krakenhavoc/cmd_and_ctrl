package game

import (
	"testing"

	"github.com/google/uuid"
)

// cleanup_discard_next_upkeep_test.go — #3490. A turn that ends on the
// active player's CR 514.1 hand-size discard goes straight into the next
// player's untap and upkeep, inside DiscardSelection. The upkeep's
// triggers (a rebound card's delayed "at the beginning of your next
// upkeep", any "at the beginning of your upkeep") were queued on
// PendingTriggers there and nobody put them on the stack: the new active
// player got priority in the upkeep over an empty stack, the table passed,
// and the triggers reached the stack in the DRAW step. CR 117.5 puts
// waiting triggers on the stack before a player receives priority, so
// they belong on the stack in the upkeep that triggered them.

// endTurnOnACleanupDiscard has the active seat end its turn the way a
// player over seven cards does: into cleanup, owing a discard, and the
// discard.
func endTurnOnACleanupDiscard(t *testing.T, g *Game) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	fillHandTo(t, g, active, 9)
	advanceIntoCleanup(t, g)
	if g.Turn.Step != StepCleanup || g.DiscardPending[active.ID] != 2 {
		t.Fatalf("expected a cleanup discard pause, at %s pending=%v", g.Turn.Step, g.DiscardPending)
	}
	picks := []uuid.UUID{active.Hand.Cards[0].InstanceID, active.Hand.Cards[1].InstanceID}
	if err := g.DiscardSelection(active.ID, picks); err != nil {
		t.Fatalf("DiscardSelection: %v", err)
	}
}

func TestUpkeepTriggersAfterACleanupDiscardGoOnTheStackInTheUpkeep(t *testing.T) {
	t.Run("a delayed next-upkeep trigger (rebound's)", func(t *testing.T) {
		g := newActiveGame(t)
		next := g.Seats[1]
		var resolvedIn Step
		fired := 0
		g.WithWriteLock(func() {
			g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
				Controller:         next.ID,
				Label:              "probe — at the beginning of your next upkeep",
				At:                 StepUpkeep,
				ControllerTurnOnly: true,
				Body: testBody(func(g *Game, _ *StackItem) error {
					fired++
					resolvedIn = g.Turn.Step
					return nil
				}),
			})
		})

		endTurnOnACleanupDiscard(t, g)

		if g.Turn.ActiveSeat != 1 || g.Turn.Step != StepUpkeep {
			t.Fatalf("after the discard: seat %d step %s, want seat 1's upkeep", g.Turn.ActiveSeat, g.Turn.Step)
		}
		if n := len(g.PendingTriggers); n != 0 {
			t.Errorf("%d triggers still waiting with seat 1 on priority, want 0 (CR 117.5)", n)
		}
		if n := len(g.StackMeta); n != 1 {
			t.Errorf("stack items in the upkeep = %d, want the delayed trigger", n)
		}
		settleStack(t, g)
		if fired != 1 || resolvedIn != StepUpkeep {
			t.Errorf("fired %d time(s), resolved in %q; want once, in the upkeep", fired, resolvedIn)
		}
	})

	t.Run("an at-the-beginning-of-your-upkeep trigger", func(t *testing.T) {
		g := newActiveGame(t)
		rec := seedUpkeepWatcher(t, g, g.Seats[1])

		endTurnOnACleanupDiscard(t, g)

		if g.Turn.ActiveSeat != 1 || g.Turn.Step != StepUpkeep {
			t.Fatalf("after the discard: seat %d step %s, want seat 1's upkeep", g.Turn.ActiveSeat, g.Turn.Step)
		}
		if n := len(g.PendingTriggers); n != 0 {
			t.Errorf("%d triggers still waiting with seat 1 on priority, want 0 (CR 117.5)", n)
		}
		settleStack(t, g)
		if rec.resolutions != 1 || rec.resolvedIn != StepUpkeep {
			t.Errorf("resolved %d time(s), in %q; want once, in the upkeep", rec.resolutions, rec.resolvedIn)
		}
	})
}
