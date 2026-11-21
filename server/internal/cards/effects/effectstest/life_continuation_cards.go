package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answerTheOrderingPrompt answers the single queued CR 616 prompt in
// the order it was offered.
func AnswerTheOrderingPrompt(t *testing.T, g *game.Game) {
	t.Helper()
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceReplacementOrder {
			prompt = c
			break
		}
	}
	if prompt == nil {
		t.Fatalf("no CR 616 ordering prompt queued (have %d choices)", len(g.PendingChoices))
	}
	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, prompt.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}
