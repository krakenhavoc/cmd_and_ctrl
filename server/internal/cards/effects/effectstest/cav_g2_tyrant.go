package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// redInPool counts the red mana in a player's pool.
func RedInPool(p *game.Player) int {
	n := 0
	for _, c := range PoolColors(p) {
		if c == "R" {
			n++
		}
	}
	return n
}
