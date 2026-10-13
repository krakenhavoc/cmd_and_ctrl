package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const gamePlanOracle = "02746b0e-3ef9-46c0-9517-6c52c2905d0b"

// TestGamePlanShufflesTheTableAndDrawsSevenThenExilesItself: every
// player's graveyard goes back into the library, each then draws seven,
// and Game Plan itself ends in exile rather than the graveyard.
func TestGamePlanShufflesTheTableAndDrawsSevenThenExilesItself(t *testing.T) {
	assistCardSpec(t, gamePlanOracle, "Game Plan")

	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushGraveyardTyped(opp, "Opt", "Instant")
	pushGraveyardTyped(me, "Ponder", "Sorcery")
	if opp.Graveyard.Size() == 0 {
		t.Fatal("setup: opponent's graveyard is empty")
	}

	id := castCatalogSpell(t, g, "Game Plan", "Sorcery", gamePlanOracle, nil)
	passPriorityAroundTable(t, g)

	if opp.Graveyard.Size() != 0 || me.Graveyard.Size() != 0 {
		t.Errorf("graveyards hold %d and %d cards, want both shuffled away", opp.Graveyard.Size(), me.Graveyard.Size())
	}
	if me.Hand.Size() != 7 || opp.Hand.Size() != 7 {
		t.Errorf("hands hold %d and %d cards, want 7 each", me.Hand.Size(), opp.Hand.Size())
	}
	if !assistCardInZone(g, id, game.ZoneExile) {
		t.Errorf("Game Plan is not in exile after it resolved")
	}
}
