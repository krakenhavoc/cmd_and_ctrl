package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func AnswerPickTargetPlayerIfOpen(t *testing.T, g *game.Game, playerID uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			if err := g.ResolvePickTarget(c.ID, c.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: playerID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			return
		}
	}
}
