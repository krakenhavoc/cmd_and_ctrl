package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trekB35Library seeds n plain cards on top of p's library.
func TrekB35Library(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		for range n {
			LibraryCardFor(p, "Filler", "Sorcery")
		}
	})
}

// trekB35SetCounters sets a battlefield card's counters of one kind.
func TrekB35SetCounters(g *game.Game, id uuid.UUID, kind string, n int) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				if g.Battlefield.Cards[i].Counters == nil {
					g.Battlefield.Cards[i].Counters = map[string]int{}
				}
				g.Battlefield.Cards[i].Counters[kind] = n
			}
		}
	})
}
