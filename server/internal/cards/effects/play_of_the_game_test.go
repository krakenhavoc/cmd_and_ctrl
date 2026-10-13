package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const playOfTheGameOracle = "337287d6-cd22-4b70-8bbb-318e3cf91bee"

// TestPlayOfTheGameExilesNonlandPermanentsAndKeepsLands: creatures on
// both sides go to exile, the caster's own included, and the land stays.
func TestPlayOfTheGameExilesNonlandPermanentsAndKeepsLands(t *testing.T) {
	assistCardSpec(t, playOfTheGameOracle, "Play of the Game")

	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 3, 3)
	plains := apaPush(g, me.ID, me.ID, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})

	castCatalogSpell(t, g, "Play of the Game", "Sorcery", playOfTheGameOracle, nil)
	passPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"my creature": mine, "their creature": theirs} {
		if !assistCardInZone(g, id, game.ZoneExile) {
			t.Errorf("%s is not in exile", name)
		}
	}
	if !assistCardOnBattlefield(g, plains) {
		t.Error("the land left the battlefield, but only nonland permanents are exiled")
	}
}
