package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func Knows(g *game.Game, p *game.Player, card, viewer uuid.UUID) bool {
	known := false
	g.ReadSnapshot(func() {
		for _, c := range p.Hand.Cards {
			if c.InstanceID == card {
				known = c.KnownBy[viewer]
			}
		}
	})
	return known
}
