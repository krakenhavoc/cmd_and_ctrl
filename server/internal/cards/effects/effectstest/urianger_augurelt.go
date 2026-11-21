package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// topCard puts one named card of the given type on top of p's library.
func TopCard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.ManaCost = cost
	p.Library.PushTop(c)
	return c.InstanceID
}
