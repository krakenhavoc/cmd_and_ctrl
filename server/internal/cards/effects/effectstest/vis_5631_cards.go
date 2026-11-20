package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// visBattlefieldNamed is the first permanent named `name` that `owner`
// controls, or false.
func VisBattlefieldNamed(g *game.Game, owner uuid.UUID, name string) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == name {
			return c, true
		}
	}
	return game.Card{}, false
}
