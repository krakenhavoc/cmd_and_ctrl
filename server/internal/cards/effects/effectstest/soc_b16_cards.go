package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// socB16TryCast puts a card in the active seat's hand and casts it, handing
// back the announce error and the card's id.
func SocB16TryCast(t *testing.T, g *game.Game, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	me.Hand.PushTop(c)
	AdvanceToMain(t, g)
	return c.InstanceID, g.CastSpell(me.ID, c.InstanceID, params)
}
