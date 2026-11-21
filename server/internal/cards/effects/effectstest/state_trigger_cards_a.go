package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func StSetLife(t *testing.T, g *game.Game, p *game.Player, life int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, life-p.Life); err != nil {
			t.Fatal(err)
		}
	})
}
