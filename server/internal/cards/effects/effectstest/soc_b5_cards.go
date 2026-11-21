package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// socB5CastX puts a catalog spell in the active seat's hand, walks to a
// main phase and casts it with the given X and targets.
func SocB5CastX(t *testing.T, g *game.Game, name, typeLine, oracle, cost string, x int, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost, Owner: me.ID, Controller: me.ID})
	AdvanceToMain(t, g)
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x, Targets: targets})
}
