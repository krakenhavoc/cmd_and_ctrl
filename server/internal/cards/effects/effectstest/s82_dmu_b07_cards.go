package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func PickTriggerTargetPlayer(t *testing.T, g *game.Game, chooser, player uuid.UUID) {
	t.Helper()
	pick := LatestPickTarget(g, chooser)
	if pick == nil {
		t.Fatal("no pending target pick")
	}
	if err := g.ResolvePickTarget(pick.ID, chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}
