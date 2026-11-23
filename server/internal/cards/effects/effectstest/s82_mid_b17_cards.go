package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func MidB17Bolt(t *testing.T, g *game.Game, target game.TargetRef) {
	t.Helper()
	B12Bolt(t, g, target)
	MonSettle(t, g)
}
