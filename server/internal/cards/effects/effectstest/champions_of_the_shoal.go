package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func ChampionsOfTheShoalStunned(t *testing.T, g *game.Game, id uuid.UUID) (tapped bool, stun int) {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped, c.Counters[game.CounterStun]
}
