package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ugins_nexus_test.go — "If a player would begin an extra turn, that
// player skips that turn instead" and "if Ugin's Nexus would be put into
// a graveyard from the battlefield, instead exile it and take an extra
// turn after this one" (ADR 0188 Decision 2, #3560, CR 614.10). Each test
// drives the real rotation seam and asserts who is active afterwards.

const (
	UginsNexusOracle      = "fe2afa18-54d0-4595-af27-256398793a42"
	HourglassPendantOracl = "01591d42-4ef8-471d-beda-d31186bba6a6"
)

func SeedNexus(g *game.Game, controller uuid.UUID) uuid.UUID {
	return SeedEnchantment(g, UginsNexusOracle, "Ugin's Nexus", "Legendary Artifact", controller)
}

func QueueOwnExtraTurn(t *testing.T, g *game.Game) {
	t.Helper()
	CastCatalogSpell(t, g, "Temporal Manipulation", "Sorcery", TemporalManipulationOracle, nil)
	PassPriorityAroundTable(t, g)
	if got := len(g.ExtraTurnsQueued()); got != 1 {
		t.Fatalf("the spell should queue its turn: %d queued", got)
	}
}

func AnswerNexusOrder(t *testing.T, g *game.Game, firstLabelContains string, wantChooser uuid.UUID) {
	t.Helper()
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceReplacementOrder {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatalf("no CR 616 ordering prompt (have %d choices)", len(g.PendingChoices))
	}
	if prompt.Chooser != wantChooser {
		t.Fatalf("the order is asked of %v, want the Nexus' controller %v", prompt.Chooser, wantChooser)
	}
	if len(prompt.ReplacementEffectIDs) != 2 {
		t.Fatalf("%d replacements in the prompt, want 2", len(prompt.ReplacementEffectIDs))
	}
	labels := map[game.ReplacementEffectID]string{}
	g.WithWriteLock(func() {
		for _, id := range prompt.ReplacementEffectIDs {
			labels[id], _ = g.ReplacementOptionMetaForEffect(id)
		}
	})
	var ids []game.ReplacementEffectID
	for id, l := range labels {
		if ContainsFold(l, firstLabelContains) {
			ids = append([]game.ReplacementEffectID{id}, ids...)
		} else {
			ids = append(ids, id)
		}
	}
	if len(ids) != 2 || !ContainsFold(labels[ids[0]], firstLabelContains) {
		t.Fatalf("no option mentioning %q in %v", firstLabelContains, labels)
	}
	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, ids); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}
