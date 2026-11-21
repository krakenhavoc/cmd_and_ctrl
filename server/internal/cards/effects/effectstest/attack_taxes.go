package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// taxTable parks a four-seat game on declare attackers with one 2/2
// under seat 0 and returns the game, the attacker and the defender.
func TaxTable(t *testing.T) (*game.Game, uuid.UUID, *game.Player) {
	t.Helper()
	g := NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	AdvanceTo(t, g, game.StepDeclareAttackers)
	return g, bear, opp
}
