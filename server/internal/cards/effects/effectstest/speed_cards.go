package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// speedGame is a catalog game walked to the active seat's precombat
// main (every max-speed sorcery ability is legal there, and a manual
// draw is not swallowed by the turn's draw step).
func SpeedGame(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceTo(t, g, game.StepPrecombatMain)
	return g, g.Seats[g.Turn.ActiveSeat]
}
