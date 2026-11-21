package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avB24Hand puts a card into p's hand and returns its ID.
func AvB24Hand(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	return c.InstanceID
}

// avB24Graveyard puts a card into p's graveyard and returns its ID.
func AvB24Graveyard(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Graveyard.PushTop(c)
	return c.InstanceID
}
