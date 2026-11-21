package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func KxPoolSize(g *game.Game, me *game.Player) int {
	n := 0
	g.WithWriteLock(func() { n = len(me.ManaPool) })
	return n
}
