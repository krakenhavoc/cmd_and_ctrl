package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rtRetap taps the permanent again, as a second source of tapping would.
func RtRetap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
}
