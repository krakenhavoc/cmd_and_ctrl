package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const cleansingOracle = "46b83657-1699-4162-a4bb-b792da396fe6"

func CleansingLand(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return SeedPermanentWithOracle(g, owner, name, "Basic Land — "+name, "")
}

// cleansingCast casts Cleansing and returns once the first leg is open.
func CleansingCast(t *testing.T, g *game.Game) {
	t.Helper()
	CastCatalogSpell(t, g, "Cleansing", "Sorcery", cleansingOracle, nil)
	PassPriorityAroundTable(t, g)
}

// cleansingAnswer answers every open leg: picks[chooser] is that player's
// choice (nil = pay for nothing).
func CleansingAnswer(t *testing.T, g *game.Game, picks map[uuid.UUID][]uuid.UUID) {
	t.Helper()
	for i := 0; i < 8; i++ {
		leg := EachOpenLeg(g)
		if leg == nil {
			return
		}
		if err := g.ResolveChooseCards(leg.ID, leg.Chooser, picks[leg.Chooser]); err != nil {
			t.Fatalf("answer for %v: %v", leg.Chooser, err)
		}
	}
	t.Fatal("legs never closed")
}
