package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mscM6Toughness is the counter-aware toughness after the layers have run.
func MscM6Toughness(g *game.Game, id uuid.UUID) int {
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c := FindBattlefieldCardByID(g, id); c != nil {
		return c.CurrentToughness()
	}
	return -1
}
