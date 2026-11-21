package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castPaid casts a plain instant of the given mana cost by paying
// exactly that much (strict), leaving it on the stack.
func CastPaid(t *testing.T, g *game.Game, me *game.Player, name, cost, pool string) {
	t.Helper()
	id := HandCardFull(me, name, "Instant", cost, "", nil)
	FloatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
}
