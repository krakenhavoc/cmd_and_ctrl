package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func AnswerPickTargetIfOpen(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			AnswerPickTarget(t, g, target)
			return
		}
	}
}
