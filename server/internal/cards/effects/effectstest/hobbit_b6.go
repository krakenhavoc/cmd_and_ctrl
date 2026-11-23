package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4454, The Hobbit (HOB / HOC) cards, half B, batch 6. Every test drives
// the registered catalog row through the engine.

// hobB6BattlefieldCount is how many permanents named `name` the player
// controls.
func HobB6BattlefieldCount(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.Name == name {
				n++
			}
		}
	})
	return n
}

// hobB6Life is a player's life total, read under the game's lock.
func HobB6Life(g *game.Game, p *game.Player) int {
	life := 0
	g.ReadSnapshot(func() { life = p.Life })
	return life
}
