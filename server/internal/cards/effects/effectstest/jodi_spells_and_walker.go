package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// battlefieldCardTapped reads the live Tapped flag.
func BattlefieldCardTapped(g *game.Game, id uuid.UUID) (bool, bool) {
	c, ok := BattlefieldCard(g, id)
	return c.Tapped, ok
}
