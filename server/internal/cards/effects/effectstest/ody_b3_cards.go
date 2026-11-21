package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// odyB3Source puts a catalog creature under the active seat in its main
// phase, not summoning sick, with `mana` in its pool.
func OdyB3Source(t *testing.T, g *game.Game, name, typeLine, oracle string, mana ...string) (*game.Player, uuid.UUID) {
	t.Helper()
	AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := PushCatalogPermanent(g, me.ID, name, typeLine, oracle, false)
	B06AddMana(me, mana...)
	return me, id
}
