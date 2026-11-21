package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// plansFor reports whether the auto-tapper finds a plan for `cost`, and
// the sources it would spend.
func PlansFor(t *testing.T, g *game.Game, me *game.Player, cost string) ([]uuid.UUID, bool) {
	t.Helper()
	parsed, err := game.ParseCost(cost)
	if err != nil {
		t.Fatalf("ParseCost(%s): %v", cost, err)
	}
	return g.AutoTapForCost(me.ID, parsed, 0)
}
