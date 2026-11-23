package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answerTriggerTargets2 answers a pick-target prompt with a player.
func AnswerTriggerTargets2(t *testing.T, g *game.Game, chooser, player uuid.UUID) {
	t.Helper()
	pick := LatestPickTarget(g, chooser)
	if pick == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTargets(pick.ID, chooser, []game.TargetRef{{Kind: game.TargetPlayer, ID: player}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}
