package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func GrantedKeysFor(g *game.Game, seat, card uuid.UUID, zone game.ZoneKind) []string {
	var out []string
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			return
		}
		for _, ac := range g.CastOffersForLocked(seat, c, zone, g.CastPermissionForLocked(seat, c, zone)) {
			if ac == nil {
				out = append(out, "printed")
			} else {
				out = append(out, ac.Key)
			}
		}
	})
	return out
}
