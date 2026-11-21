package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dftCG2Tapped reports whether a battlefield permanent is tapped.
func DftCG2Tapped(g *game.Game, id uuid.UUID) bool {
	c, ok := BattlefieldCard(g, id)
	return ok && c.Tapped
}
