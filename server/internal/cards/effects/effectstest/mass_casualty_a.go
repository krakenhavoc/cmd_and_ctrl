package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func McPoolCount(p *game.Player, color string) int {
	n := 0
	for _, tok := range p.ManaPool.Units() {
		if tok.Color == color {
			n++
		}
	}
	return n
}
