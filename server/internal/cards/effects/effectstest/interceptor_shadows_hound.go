package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func InterceptorOnBattlefield(g *game.Game, id uuid.UUID) (game.Card, bool) {
	var c game.Card
	var ok bool
	g.ReadSnapshot(func() {
		if g.Battlefield.Contains(id) {
			c, ok = g.LookupCardForEffect(id)
		}
	})
	return c, ok
}
