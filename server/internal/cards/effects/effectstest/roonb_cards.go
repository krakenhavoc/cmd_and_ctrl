package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func RoonbControllerOf(g *game.Game, id uuid.UUID) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Controller
		}
	}
	return uuid.Nil
}
