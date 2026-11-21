package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func FindRestoredCard(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	t.Fatalf("restored battlefield lost card %s", id)
	return nil
}
