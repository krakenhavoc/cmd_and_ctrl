package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const playOfTheGameOracle = "337287d6-cd22-4b70-8bbb-318e3cf91bee"

// TestPlayOfTheGameExilesNonlandPermanentsAndKeepsLands: creatures on
// both sides go to exile, the caster's own included, and the land stays.
func TestPlayOfTheGameExilesNonlandPermanentsAndKeepsLands(t *testing.T) {
	effectstest.AssistCardSpec(t, playOfTheGameOracle, "Play of the Game")

	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 3, 3)
	plains := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})

	effectstest.CastCatalogSpell(t, g, "Play of the Game", "Sorcery", playOfTheGameOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"my creature": mine, "their creature": theirs} {
		if !effectstest.AssistCardInZone(g, id, game.ZoneExile) {
			t.Errorf("%s is not in exile", name)
		}
	}
	if !effectstest.AssistCardOnBattlefield(g, plains) {
		t.Error("the land left the battlefield, but only nonland permanents are exiled")
	}
}
