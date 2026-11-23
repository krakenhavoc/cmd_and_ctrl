package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// treasureCount counts the Treasure tokens on the battlefield.
func TreasureCount(g *game.Game) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Treasure") {
			n++
		}
	}
	return n
}
