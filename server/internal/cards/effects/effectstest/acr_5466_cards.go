package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func Acrb1AddMana(t *testing.T, g *game.Game, player uuid.UUID, mana string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(player, uuid.Nil, mana); err != nil {
			t.Fatalf("mana: %v", err)
		}
	})
}
