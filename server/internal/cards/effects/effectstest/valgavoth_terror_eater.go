package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func CardInExile(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	c, ok := ExileCard(g, id)
	if !ok {
		t.Fatalf("card %s is not in exile", id)
	}
	return c
}
