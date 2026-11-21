package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func GraveyardPermission(g *game.Game, player, card uuid.UUID) *game.CastPermission {
	var perm *game.CastPermission
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			return
		}
		perm = g.CastPermissionForLocked(player, c, game.ZoneGraveyard)
	})
	return perm
}
