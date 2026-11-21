package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ecl2CastAt puts a catalog card in the active player's hand and casts it
// in the current step, with no mana check.
func Ecl2CastAt(t *testing.T, g *game.Game, name, typeLine, oracle string, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets})
}
