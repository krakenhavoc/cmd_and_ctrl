package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// copyTheSpell puts a copy of the spell `spell` on the stack under `me`.
func CopyTheSpell(t *testing.T, g *game.Game, me *game.Player, spell uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(spell, me.ID, false, nil); err != nil {
			t.Fatalf("copy: %v", err)
		}
	})
}
