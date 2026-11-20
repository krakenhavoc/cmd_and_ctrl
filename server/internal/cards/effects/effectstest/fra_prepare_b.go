package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fraBCastCopy casts the prepare spell's copy out of exile.
func FraBCastCopy(t *testing.T, g *game.Game, p *game.Player, spell string, targets ...game.TargetRef) {
	t.Helper()
	cp, ok := PrepareCopyOf(g, spell)
	if !ok {
		t.Fatalf("no copy of %s in exile", spell)
	}
	if err := g.CastSpell(p.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile), Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", spell, err)
	}
	PassPriorityAroundTable(t, g)
}
