package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func ColorlessFloating(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = p.ManaPool.CountColor("C") })
	return n
}
