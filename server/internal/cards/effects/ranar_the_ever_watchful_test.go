package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ranar_the_ever_watchful_test.go — #1319, against the REAL catalog:
// Ranar's foretell discount, exercised on real foretell cards (Saw It
// Coming, Behold the Multiverse) rather than a fixture, because the
// point of the proof is that a printed permanent and a printed
// foretell card compose correctly through the shared CR 601.2f pass.
//
// ranarOracle is stack_exile_cards_test.go's (#1320) — one constant,
// shared, rather than a second copy of the same UUID string.

// TestRanarMakesTheFirstForetellEachTurnFreeButNotTheSecond is the
// end-to-end shape, on the PAYING path: STRICT mode with an empty
// pool is the whole assertion — permissive mode would wave an
// unaffordable {2} through with a cost warning and this would pass
// whether or not the discount actually priced anything.
func TestRanarMakesTheFirstForetellEachTurnFreeButNotTheSecond(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	pushCatalogPermanent(g, active.ID, "Ranar the Ever-Watchful", "Legendary Creature — Spirit Warrior", ranarOracle, false)

	first := handCardForTest(active, "Saw It Coming", "Instant", sawItComingOracle)
	second := handCardForTest(active, "Behold the Multiverse", "Instant", beholdTheMultiverseOracl)

	if err := g.PerformSpecialAction(active.ID, first, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("first foretell this turn (should be free under Ranar): %v", err)
	}
	if len(active.ManaPool) != 0 {
		t.Errorf("mana pool after the free foretell: got %d tokens, want 0", len(active.ManaPool))
	}
	if got := g.ForetoldCountThisTurn(active.ID); got != 1 {
		t.Fatalf("ForetoldCountThisTurn after one foretell: got %d, want 1", got)
	}

	// "The FIRST card", not "the first N": the second foretell this
	// turn costs the printed {2}.
	active.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(active.ID, second, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("second foretell this turn (should cost the printed {2}, and the pool can pay it): %v", err)
	}
	if len(active.ManaPool) != 0 {
		t.Errorf("mana pool after the second foretell: got %d tokens, want 0 — the printed {2} should have been charged", len(active.ManaPool))
	}
}

// TestRanarsDiscountDoesNotReachAnOrdinaryForetellWithNoRanar is the
// partition's other half, pinned on this card specifically: remove
// Ranar from the board and the same cards cost the printed {2} again.
func TestRanarsDiscountDoesNotReachAnOrdinaryForetellWithNoRanar(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	card := handCardForTest(active, "Saw It Coming", "Instant", sawItComingOracle)

	if err := g.PerformSpecialAction(active.ID, card, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("foretold with an empty pool and no Ranar on the board — the printed {2} should have refused it")
	}
}
