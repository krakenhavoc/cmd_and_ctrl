package effectstest

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// damage_continuation_cards_test.go is #807 seen from the catalog:
// Creeping Bloodsucker — "this creature deals 1 damage to each
// opponent. You gain life equal to the damage dealt this way" — on a
// board where the damage dealt this way is NOT 1 × the number of
// opponents.
//
// The engine-side suite is server/internal/game/damage_continuation_test.go.
// What this test adds is the card sentence, on a board built entirely
// out of real cards and real primitives: Angrath's Marauders doubles
// every damage event from a source its controller controls, and a
// CR 615.8 prevention shield sits on one opponent. Two DIFFERENT damage
// replacements on that one opponent is the mixed window that still
// prompts after #800 — identical replacements collapse without asking,
// and this pair is not identical.
//
// Before #807 the card read each opponent's life total back on the line
// after damaging them. The shielded opponent's damage was still waiting
// on the CR 616 ordering prompt, so their total had not moved, they
// counted as having taken nothing, and the gain was short by everything
// they took.

const B14AngrathsMaraudersOracle = "2d4976d4-649c-4d42-ac5a-ada4b46a480c"

// answerOrderingPromptDoublerFirst answers the single queued CR 616
// ordering prompt with the damage doubler applied FIRST.
//
// The order is chosen rather than taken as offered because it is the
// order that makes the test say something: doubler-then-shield leaves 1
// damage for the continuation to report, while shield-then-doubler
// prevents the whole 1 and reports zero — which is also what the bug
// reported, so it could not tell the two apart.
func AnswerOrderingPromptDoublerFirst(t *testing.T, g *game.Game) {
	t.Helper()
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceReplacementOrder {
			prompt = c
			break
		}
	}
	if prompt == nil {
		t.Fatalf("no CR 616 ordering prompt queued (have %d choices) — "+
			"a doubler and a prevention shield on one event are two different replacements",
			len(g.PendingChoices))
	}
	if len(prompt.ReplacementEffectIDs) != 2 {
		t.Fatalf("prompt offers %d replacements, want 2", len(prompt.ReplacementEffectIDs))
	}
	ordered := make([]game.ReplacementEffectID, 0, 2)
	var shield []game.ReplacementEffectID
	for _, id := range prompt.ReplacementEffectIDs {
		label, _ := g.ReplacementOptionMetaForEffect(id)
		if strings.Contains(strings.ToLower(label), "prevent") {
			shield = append(shield, id)
			continue
		}
		ordered = append(ordered, id)
	}
	if len(shield) != 1 || len(ordered) != 1 {
		t.Fatalf("expected exactly one prevention shield and one doubler in the prompt, got %d/%d",
			len(shield), len(ordered))
	}
	ordered = append(ordered, shield...)
	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, ordered); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}
