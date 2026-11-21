package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func GainLifeNow(t *testing.T, g *game.Game, source uuid.UUID, p *game.Player, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(source, p.ID, n); err != nil {
			t.Fatal(err)
		}
	})
}
