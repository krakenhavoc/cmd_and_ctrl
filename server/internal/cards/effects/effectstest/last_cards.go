package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Training Grounds ---------------------------------------------------------

// lcPrice prices a made-up ability of `cost` on the permanent `id` for
// `activator`, at X = x, through the same pricer the activation pays.
func LcPrice(t *testing.T, g *game.Game, activator, id uuid.UUID, cost string, x int) string {
	t.Helper()
	var (
		q   game.ParsedCost
		err error
	)
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("no card %s", id)
		}
		ab := game.ActivatedAbilityShape{Label: cost + ": test", Cost: game.AbilityCost{Mana: cost}}
		q, err = g.AbilityManaCostAtXForEffect(activator, c, game.ZoneBattlefield, ab, nil, x)
	})
	if err != nil {
		t.Fatalf("price %s: %v", cost, err)
	}
	return q.SettleX(x).String()
}
