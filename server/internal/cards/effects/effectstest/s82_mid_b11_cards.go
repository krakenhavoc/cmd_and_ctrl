package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func S82MidB11Flashback(g *game.Game, id uuid.UUID, params game.CastSpellParams) error {
	me := g.Seats[g.Turn.ActiveSeat]
	params.FromZone, params.AlternativeCost, params.Strict = "graveyard", "flashback", true
	return g.CastSpell(me.ID, id, params)
}
