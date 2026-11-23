package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func KrileConfirm(t *testing.T, g *game.Game, me *game.Player, yes bool) {
	t.Helper()
	ask := LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatalf("expected Krile's confirm prompt, got %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, yes); err != nil {
		t.Fatalf("ResolveConfirm(%v): %v", yes, err)
	}
	PassPriorityAroundTable(t, g)
}
