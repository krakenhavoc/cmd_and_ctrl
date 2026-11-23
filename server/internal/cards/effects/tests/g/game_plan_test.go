package testsg

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const gamePlanOracle = "02746b0e-3ef9-46c0-9517-6c52c2905d0b"

// TestGamePlanShufflesTheTableAndDrawsSevenThenExilesItself: every
// player's graveyard goes back into the library, each then draws seven,
// and Game Plan itself ends in exile rather than the graveyard.
func TestGamePlanShufflesTheTableAndDrawsSevenThenExilesItself(t *testing.T) {
	effectstest.AssistCardSpec(t, gamePlanOracle, "Game Plan")

	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushGraveyardTyped(opp, "Opt", "Instant")
	effectstest.PushGraveyardTyped(me, "Ponder", "Sorcery")
	if opp.Graveyard.Size() == 0 {
		t.Fatal("setup: opponent's graveyard is empty")
	}

	id := effectstest.CastCatalogSpell(t, g, "Game Plan", "Sorcery", gamePlanOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Graveyard.Size() != 0 || me.Graveyard.Size() != 0 {
		t.Errorf("graveyards hold %d and %d cards, want both shuffled away", opp.Graveyard.Size(), me.Graveyard.Size())
	}
	if me.Hand.Size() != 7 || opp.Hand.Size() != 7 {
		t.Errorf("hands hold %d and %d cards, want 7 each", me.Hand.Size(), opp.Hand.Size())
	}
	if !effectstest.AssistCardInZone(g, id, game.ZoneExile) {
		t.Errorf("Game Plan is not in exile after it resolved")
	}
}
