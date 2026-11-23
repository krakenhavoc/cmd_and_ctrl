package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TriggerOnStackAny(g *game.Game) bool {
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == game.StackItemTriggered {
			return true
		}
	}
	return false
}
