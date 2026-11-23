package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hand_revealed.go — revealed-hand test helpers shared across test packages.

func HrRevealed(g *game.Game, player uuid.UUID) []string {
	var out []string
	g.ReadSnapshot(func() { out = g.HandRevealedLocked(player) })
	return out
}
