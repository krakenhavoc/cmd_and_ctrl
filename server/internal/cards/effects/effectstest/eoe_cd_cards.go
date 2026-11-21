package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- helpers ---------------------------------------------------------------

// eoeCDCard is a target ref to a card.
func EoeCDCard(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// eoeCDPT is a permanent's power and toughness after the layers, with its
// counters.
func EoeCDPT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.PowerForComparison(), c.CurrentToughness()
}

// eoeCDSet edits a battlefield permanent in place under the write lock.
func EoeCDSet(g *game.Game, id uuid.UUID, edit func(c *game.Card)) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				edit(&g.Battlefield.Cards[i])
			}
		}
	})
}

// eoeCDActivate activates a catalog row strictly, paying from the pool.
func EoeCDActivate(g *game.Game, who, source uuid.UUID, idx int, targets ...game.TargetRef) error {
	return g.ActivateCatalogAbility(who, source, idx, game.ActivateAbilityParams{Strict: true, Targets: targets})
}

// eoeCDCount counts the permanents a player controls with a name.
func EoeCDCount(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}
