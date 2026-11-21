package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Winota, Joiner of Forces ---------------------------------------------------------

func FinB12AnswerAttackPick(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser == who {
			if err := g.ResolveOptionPick(c.ID, who, 0); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			return
		}
	}
	t.Fatal("no \"what does it attack?\" prompt")
}
