package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// radDestroy destroys a permanent and settles whatever it triggered
// down to a prompt.
func RadDestroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id, game.DestroyOptions{}); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
}
