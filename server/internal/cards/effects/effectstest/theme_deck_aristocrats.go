package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// battlefieldIDNamed returns the first battlefield card a player
// controls with the given name.
func BattlefieldIDNamed(g *game.Game, controller uuid.UUID, name string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			return c.InstanceID
		}
	}
	return uuid.Nil
}
