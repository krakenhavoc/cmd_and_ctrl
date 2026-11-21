package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_a_goliath_test.go — Goliath Daydreamer (#3213) and the engine
// piece under it, a spell exiled with a counter as it resolves.

func GoliathGame(t *testing.T) (*game.Game, uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[0]
	return g, PushDiesCreatureForTest(g, me.ID, "Goliath Daydreamer", GoliathOracle, "Creature — Giant Wizard", 4, 4)
}

func ExiledCard(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == id {
			return &g.Exile.Cards[i]
		}
	}
	return nil
}
