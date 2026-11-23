package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// socB2Token seeds a creature token (a "Token" type line) that can attack.
func SocB2Token(g *game.Game, owner uuid.UUID, name, subtype string, power, toughness int) uuid.UUID {
	return B12Creature(g, owner, name, "Token Creature — "+subtype, power, toughness)
}
