package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// windowAt reports whether the exiled card's grant is live for `me`.
func WindowAt(g *game.Game, id uuid.UUID, me uuid.UUID) bool {
	perm := g.CastPermissionOnCardByIDForEffect(id)
	return perm != nil && PermissionLive(g, perm, me)
}
