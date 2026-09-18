package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// cleanup_priority_test.go — #661 / CR 514.3a from the enumerator's
// side. The cleanup step normally grants nobody priority, and the bot
// runner and the client's auto-pass both read "what can this seat do"
// off this enumerator rather than off the step. So the one thing that
// has to be true of the new window is that it looks like every other
// priority window: the holder is offered a pass, and nobody wedges.

// TestCleanupPriorityWindowOffersAPass sets up the CR 514.3a
// condition with no trigger at all — a state-based action caused by
// the CR 514.2 sweep itself. A 2/2 kept alive by an "until end of
// turn" +3/+3 with four -1/-1 counters on it has 0 or less toughness
// the instant the pump ends, so the check in the cleanup step
// performs an SBA and the active player gets priority there.
func TestCleanupPriorityWindowOffersAPass(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	// No hand-size discard: the SBA is the only thing happening in
	// this cleanup step, so the window it opens is the thing measured.
	clearHand(active)
	bear := battlefieldCard(g, active, creature("Doomed Bear", "{1}{G}", 2, 2))
	g.WithWriteLock(func() {
		g.RegisterScopedStaticForEffect(game.StaticAbility{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.InstanceID == bear
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 3
				c.Toughness += 3
			},
		}, uuid.New(), "test — +3/+3 until end of turn", g.UntilEndOfTurnDuration())
	})
	if err := g.AddCounter(bear, "-1/-1", 4); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}

	advanceTo(t, g, game.StepEnd)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep past End: %v", err)
	}
	if g.Turn.Step != game.StepCleanup {
		t.Fatalf("no CR 514.3a window: step %s", g.Turn.Step)
	}

	moves := legal.EnumerateFor(g, active.ID)
	if !hasLabel(moves, "Pass priority") {
		t.Fatalf("the seat holding priority in cleanup was offered no pass: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)

	// Everyone passing walks the table out of the window and on to the
	// next turn — the bot runner's loop, in miniature.
	for i := 0; i < 8 && g.Turn.Step == game.StepCleanup; i++ {
		holder := g.Seats[g.Turn.PriorityHolder]
		if holder == nil {
			t.Fatalf("nobody holds priority inside the CR 514.3a window")
		}
		if !hasLabel(legal.EnumerateFor(g, holder.ID), "Pass priority") {
			t.Fatalf("%s holds priority in cleanup with no pass offered", holder.Name)
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if g.Turn.Step == game.StepCleanup {
		t.Fatalf("the table never got out of the cleanup window")
	}
	if g.Turn.ActiveSeat == 0 {
		t.Errorf("the turn did not end after the second cleanup step")
	}
}
