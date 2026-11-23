package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// handCardByID returns a card from a seat's hand by instance ID.
func HandCardByID(g *game.Game, p *game.Player, id uuid.UUID) (game.Card, bool) {
	var out game.Card
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range p.Hand.Cards {
			if c.InstanceID == id {
				out, found = c, true
				return
			}
		}
	})
	return out, found
}
