package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ecl5Card reads a battlefield card, failing the test if it is gone.
func Ecl5Card(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c
}
