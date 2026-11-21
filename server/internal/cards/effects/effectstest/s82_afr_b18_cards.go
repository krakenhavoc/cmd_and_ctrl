package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func B04WaitForPickOrSkip(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	if LatestPickTarget(g, chooser) != nil {
		PickCard(t, g, chooser, target)
	}
}
