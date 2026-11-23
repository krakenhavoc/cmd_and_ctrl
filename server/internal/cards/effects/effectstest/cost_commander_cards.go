package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cost_commander_cards_test.go — #1397's proof cards: five catalog
// cards whose COST can move a commander, one per cost component the
// fix touched. Each one used to settle without asking (the discard,
// return and exile payers) or pause half way through the payment (the
// sacrifice). The engine half — every component, both answers, undo —
// is game/cost_commander_choice_test.go.
//
// Since ADR 0115 only a cost that puts a card into a HAND or a LIBRARY
// (CR 903.9b) parks the announcement on a question to the commander's
// owner and pays nothing until it is answered: ninjutsu below. A commander
// discarded, exiled or sacrificed to pay a cost is paid like any other
// card, and the CR 903.9a state-based action asks its owner afterwards;
// the other four cards here pin that order.

// costCommander builds a commander card owned by `owner`.
func CostCommander(owner uuid.UUID, name string) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Creature — Human Wizard",
		Power: 3, Toughness: 3, Owner: owner, Controller: owner, IsCommander: true,
		KnownBy: map[uuid.UUID]bool{owner: true},
	}
}

// answerCostCommander answers the one open CR 903.9b prompt as `owner`,
// after checking it is addressed to them and nothing has been paid.
func AnswerCostCommander(t *testing.T, g *game.Game, owner uuid.UUID, apply bool) {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the one CR 903.9 prompt", len(g.PendingChoices))
	}
	c := g.PendingChoices[0]
	if c.Kind != game.PendingChoiceOptionalReplacement || c.Chooser != owner {
		t.Fatalf("prompt %q owed by %s, want an optional_replacement owed by the commander's owner %s", c.Kind, c.Chooser, owner)
	}
	if err := g.ResolveOptionalReplacement(c.ID, owner, apply); err != nil {
		t.Fatalf("answer: %v", err)
	}
}
