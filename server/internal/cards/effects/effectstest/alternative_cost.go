package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castWithAltCost is castCatalogSpell with an alternative cost
// claimed. Seeds the card in the active seat's hand, walks to a main
// phase, and announces.
func CastWithAltCost(
	t *testing.T,
	g *game.Game,
	name, typeLine, oracleID, altCost string,
) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		AlternativeCost: altCost,
	}); err != nil {
		t.Fatalf("CastSpell %s (%s): %v", name, altCost, err)
	}
	return id
}

// answerAnyTriggerOrderPrompt resolves an outstanding CR 603.3b
// ordering prompt for `chooser`, keeping the queue's current order.
// Returns false when there is no prompt waiting. Bounded priority
// passes first, because the prompt only appears once the drain runs.
func AnswerAnyTriggerOrderPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) bool {
	t.Helper()
	var prompt *game.PendingChoice
	for i := 0; i < 8 && prompt == nil; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == chooser {
				prompt = c
			}
		}
		if prompt != nil {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if prompt == nil {
		return false
	}
	// Answer with exactly the set the prompt covers, in the order it
	// offered — ResolveTriggerOrder rejects anything else.
	order := append([]uuid.UUID(nil), prompt.TriggerOrderIDs...)
	if err := g.ResolveTriggerOrder(prompt.ID, chooser, order); err != nil {
		t.Fatalf("ResolveTriggerOrder: %v", err)
	}
	return true
}

// passPriorityUntilChoice passes priority until some seat is owed a
// pending choice, the stack is empty, or the table has gone around
// too many times. Tests that expect a prompt mid-resolution need it:
// passPriorityAroundTable drains the stack unconditionally, which
// walks straight past an unanswered prompt (PassPriority does not
// gate on open choices — see spell_copy.go).
func PassPriorityUntilChoice(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if len(g.PendingChoices) > 0 || StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatal("no choice was queued and the stack did not empty after 32 priority passes")
}
