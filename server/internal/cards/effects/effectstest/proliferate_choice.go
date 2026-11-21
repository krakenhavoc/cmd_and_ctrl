package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// proliferate_choice_test.go — #2525: proliferate is the player's
// choice (CR 701.34a), not an auto-pick.

// castSteadyProgressToPrompt casts Steady Progress and passes until
// its proliferate prompt is open, leaving the draw undrawn.
func CastSteadyProgressToPrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	CastCatalogSpell(t, g, "Steady Progress", "Instant", SteadyProgressOracle, nil)
	PassPriorityAroundTable(t, g)
	c := ProliferatePrompt(g)
	if c == nil {
		t.Fatal("Steady Progress did not ask what to proliferate")
	}
	return c
}
