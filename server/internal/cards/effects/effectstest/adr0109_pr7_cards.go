package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// p7Table is a catalog game in the active seat's main phase with empty
// hands and graveyards for the two seats a test uses.
func P7Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g, me, opp := ExileCostTable(t)
	g.WithWriteLock(func() {
		opp.Hand.Cards = nil
		opp.Graveyard.Cards = nil
	})
	return g, me, opp
}

// p7Activate activates row `index` of `source` for p, paid on paper,
// and resolves it.
func P7Activate(t *testing.T, g *game.Game, p *game.Player, source uuid.UUID, index int, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(p.ID, source, index, params); err != nil {
		t.Fatalf("%s activates row %d: %v", p.Name, index, err)
	}
	PassPriorityAroundTable(t, g)
}

func P7Countered(g *game.Game, spell uuid.UUID) bool {
	for _, ev := range g.Events {
		if ev.Kind == game.EventCounterSpell && ev.CardID == spell {
			return true
		}
	}
	return false
}
