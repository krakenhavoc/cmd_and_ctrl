package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ceOnBattlefield reads a permanent back by ID.
func CeOnBattlefield(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c
		}
	}
	t.Fatalf("%v is not on the battlefield", id)
	return game.Card{}
}
