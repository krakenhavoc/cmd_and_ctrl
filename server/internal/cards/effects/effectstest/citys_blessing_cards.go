package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// grantBlessing gives a player the city's blessing the way the engine
// does (the designation and its event, which also invalidates the layer
// pass), without a board of ten permanents.
func GrantBlessing(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() {
		p.CitysBlessing = true
		g.EmitEvent(game.Event{Kind: game.EventCitysBlessing, Actor: p.ID})
	})
}
