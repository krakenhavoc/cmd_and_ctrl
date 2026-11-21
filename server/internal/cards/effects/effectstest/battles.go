package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func InExile(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func AnswerProtectorPrompt(t *testing.T, g *game.Game, chooser, protector uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseProtector && c.Chooser == chooser {
			if err := g.ResolveChooseProtector(c.ID, chooser, protector); err != nil {
				t.Fatalf("ResolveChooseProtector: %v", err)
			}
			return
		}
	}
	t.Fatal("no protector prompt was queued")
}

func AnswerFirstPickTarget(t *testing.T, g *game.Game) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoicePickTarget {
			continue
		}
		if len(c.PickTargetCards) == 0 {
			t.Fatalf("pick_target prompt %q offered no cards", c.Reason)
		}
		pick := game.TargetRef{Kind: game.TargetCard, ID: c.PickTargetCards[0]}
		if err := g.ResolvePickTarget(c.ID, c.Chooser, pick); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		return
	}
	t.Fatal("no pick_target prompt was queued")
}
