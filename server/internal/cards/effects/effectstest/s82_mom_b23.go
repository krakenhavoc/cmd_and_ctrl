package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b23Library replaces a player's library with cs, the first card listed on
// top, and returns their ids in that order.
func B23Library(p *game.Player, cs ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	var ids []uuid.UUID
	for _, c := range cs {
		c.Owner, c.Controller = p.ID, p.ID
		if c.InstanceID == uuid.Nil {
			c.InstanceID = uuid.New()
		}
		ids = append(ids, PushLibraryCardForTest(p, c))
	}
	return ids
}
