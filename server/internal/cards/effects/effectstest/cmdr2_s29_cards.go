package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s29Bump makes the layers re-read the graveyard after a test seeds it.
func S29Bump(g *game.Game) { g.WithWriteLock(func() { g.BumpLayerVersionForTest() }) }

func S29Size(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return EffectivePower(t, g, id), EffectiveToughness(t, g, id)
}
