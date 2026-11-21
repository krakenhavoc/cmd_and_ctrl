package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func HobA8Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return RingBFCard(t, g, id).Counters[game.CounterPlusOne]
}
