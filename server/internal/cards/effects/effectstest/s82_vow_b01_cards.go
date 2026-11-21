package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func Vow1Pick(t *testing.T, g *game.Game, me *game.Player, ref []game.TargetRef) {
	t.Helper()
	p := LatestPickTarget(g, me.ID)
	if p == nil {
		PassPriorityAroundTable(t, g)
		p = LatestPickTarget(g, me.ID)
	}
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, ref); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	PassPriorityAroundTable(t, g)
}
