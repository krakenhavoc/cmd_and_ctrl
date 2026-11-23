package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shuffleCount is how many times `player` shuffled their own library.
func ShuffleCount(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Actor == player && ev.Label == "shuffle" {
			n++
		}
	}
	return n
}

func LibraryHas(p *game.Player, name string) bool {
	for _, c := range p.Library.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}
