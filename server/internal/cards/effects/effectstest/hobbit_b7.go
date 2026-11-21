package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func HobB7Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Counters[game.CounterPlusOne]
}
