package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_a_fandaniel_test.go — Fandaniel, Telophoroi Ascian (#3213).

func FandanielGame(t *testing.T, spellsInGraveyard int) *game.Game {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[0]
	PushDiesCreatureForTest(g, me.ID, "Fandaniel, Telophoroi Ascian", fandanielOracle, "Legendary Creature — Elder Wizard", 4, 5)
	for i := 0; i < spellsInGraveyard; i++ {
		typ := "Instant"
		if i%2 == 1 {
			typ = "Sorcery"
		}
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spent Spell", TypeLine: typ, Owner: me.ID, Controller: me.ID})
	}
	// A creature card in the graveyard does not count.
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	return g
}
