package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Confusticate and Bebother ---------------------------------------------

// hobA7BearSpell puts a creature spell of mine on the stack.
func HobA7BearSpell(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	ToMainForCost(t, g)
	bear := HandCardFull(me, "Bear Cub", "Creature — Bear", "", "", []string{"G"})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast creature: %v", err)
	}
	return bear
}
