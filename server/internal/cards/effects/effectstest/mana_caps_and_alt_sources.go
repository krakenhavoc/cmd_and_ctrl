package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushTreasure puts an untapped Treasure token under p's control.
func PushTreasure(g *game.Game, p *game.Player) uuid.UUID {
	tok := effects.TreasureToken()
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = p.ID, p.ID
	g.WithWriteLock(func() { g.Battlefield.PushTop(tok) })
	return tok.InstanceID
}
