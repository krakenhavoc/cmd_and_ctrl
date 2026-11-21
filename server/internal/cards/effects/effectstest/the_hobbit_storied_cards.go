package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// giveStory gives a player an enduring story the way the engine does
// (the flag and its event, which invalidates the layer pass), without a
// board of three artifacts.
func GiveStory(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() {
		p.EnduringStory = true
		g.EmitEvent(game.Event{Kind: game.EventEnduringStory, Actor: p.ID})
	})
}
