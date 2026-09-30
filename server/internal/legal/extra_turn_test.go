package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// TestLoyaltyActivationIsOfferedAgainInAnExtraTurn is ADR 0059 test 19's
// second half: CR 606.3 is once per TURN, and an extra turn is a turn
// (CR 500.7), so the same seat taking two turns in a row gets its
// loyalty activation back — and the offered move is one the engine
// accepts (#544).
func TestLoyaltyActivationIsOfferedAgainInAnExtraTurn(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	pw := walker(g, active, "Karn Liberated", oracleKarn, 8)
	advanceTo(t, g, game.StepPrecombatMain)
	// Ability 0 is the +4 ("target player exiles a card from their
	// hand"); the −3 would exile Karn himself.
	if err := g.ActivateCatalogAbility(active.ID, pw, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: active.ID}},
	}); err != nil {
		t.Fatalf("+4: %v", err)
	}
	for i := 0; i < 16 && (g.Stack.Size() > 0 || len(g.PendingTriggers) > 0); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if n := len(activationsOf(legal.EnumerateFor(g, active.ID), pw)); n != 0 {
		t.Fatalf("setup: %d activations still offered this turn", n)
	}

	g.WithWriteLock(func() { g.TakeExtraTurnsForEffect(active.ID, uuid.Nil, 1) })
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if g.Turn.ActiveSeat != active.Seat || !g.Turn.Extra {
		t.Fatalf("expected %s's extra turn, got seat %d extra=%v", active.Name, g.Turn.ActiveSeat, g.Turn.Extra)
	}
	advanceTo(t, g, game.StepPrecombatMain)
	acts := activationsOf(legal.EnumerateFor(g, active.ID), pw)
	if len(acts) == 0 {
		t.Fatal("no loyalty activation offered in the extra turn")
	}
	dispatchAll(t, g, active.ID, acts)
}
