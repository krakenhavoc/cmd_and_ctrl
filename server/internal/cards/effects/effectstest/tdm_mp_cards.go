package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tdmMPCastFreeSpell casts a {0} sorcery from `p`'s hand and resolves it.
func TdmMPCastFreeSpell(t *testing.T, g *game.Game, p *game.Player, name string) {
	t.Helper()
	id := MayhemHandCard(g, p, name, "Sorcery", "{0}", "")
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
}
