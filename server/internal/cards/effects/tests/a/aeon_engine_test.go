package testsa

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aeon_engine_test.go — Aeon Engine (#6223): it enters tapped, and its
// ability, paid by tapping and exiling it, reverses the game's turn
// order (CR 103.1, CR 101.4).

const aeonEngineOracle = "af471455-6b9f-4c14-bb90-a74a44ecd247"

func TestAeonEngineEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	engine := effectstest.CastCatalogSpell(t, g, "Aeon Engine", "Artifact", aeonEngineOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, engine) {
		t.Fatal("Aeon Engine entered untapped")
	}
	if err := g.ActivateCatalogAbility(g.Seats[0].ID, engine, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a tapped Aeon Engine paid its {T} cost")
	}
}

// Seats A-B-C-D; A activates on A's turn. The engine is exiled as the
// cost, and once the ability resolves the next turn is D's, then C's.
func TestAeonEngineReversesTheTurnOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if g.Turn.ActiveSeat != me.Seat {
		t.Fatalf("setup: seat %d starts", g.Turn.ActiveSeat)
	}
	engine := effectstest.B12Push(g, me.ID, "Aeon Engine", "Artifact", aeonEngineOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, engine, 0, game.ActivateAbilityParams{})
	if !g.Exile.Contains(engine) || g.Battlefield.Contains(engine) {
		t.Fatal("Aeon Engine is not in exile once its cost is paid")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.TurnOrderReversedForEffect() {
		t.Fatal("the ability resolved without reversing the turn order")
	}
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, 3, false, "the turn after the reversal")
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, 2, false, "the turn after that")
}
