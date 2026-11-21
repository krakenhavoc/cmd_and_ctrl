package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func BolsterCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return FlipState(t, g, id).Counters[game.CounterPlusOne]
}
