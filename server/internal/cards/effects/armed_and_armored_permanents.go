package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// armed_and_armored_permanents.go — a player's matching permanents, shared
// by Armed and Armored and Battlefield Improvisation.

// ArmedAndArmoredPermanents lists the battlefield permanents a player
// controls that match.
func ArmedAndArmoredPermanents(g *game.Game, controller uuid.UUID, match func(game.Card) bool) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && match(c) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
