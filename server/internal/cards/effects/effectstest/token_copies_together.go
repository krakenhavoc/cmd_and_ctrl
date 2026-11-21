package effectstest

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_copies_together_test.go — #4594: a "for each creature, create a
// token that's a copy of it" effect creates every copy at the same time,
// in one creation (CR 608.2f), so the copies enter the battlefield in
// one event and each newcomer is checked for the others' entry
// (CR 603.6a). The probe is exile_return_together_test.go's: a Soul
// Warden ("whenever another creature enters, you gain 1 life") token
// copy that enters with another sees it enter; one created before the
// other does not. Mister Fantastic's "whenever one or more tokens you
// control enter" is the batch trigger: one event, one trigger
// (CR 603.2c).

// tctSoulWarden puts a Soul Warden under p's control.
func TctSoulWarden(g *game.Game, p uuid.UUID) uuid.UUID {
	return B27Push(g, p, "Soul Warden", "Creature — Human Cleric", B04SoulWardenOracle, "{W}", 1, 1, "W")
}

// tctMisterFantastic puts Mister Fantastic under p's control.
func TctMisterFantastic(g *game.Game, p uuid.UUID) uuid.UUID {
	return B27Push(g, p, "Mister Fantastic, Reed Richards", "Legendary Creature — Human Scientist Hero", MshA5FantasticOracle, "{3}{U}", 2, 4, "U")
}

// tctSettle resolves the stack, saying yes to every optional trigger
// chooser is asked about and taking every trigger order as offered. A
// dozen triggers is more than assistCardSettle and passPriorityAroundTable
// are bounded for, so it passes priority one seat at a time.
func TctSettle(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 256; i++ {
		if c := AssistCardTriggerPrompt(g, chooser); c != nil {
			if err := g.ResolveTriggerPrompt(c.ID, c.Chooser, true); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			continue
		}
		if TriggerOrderPrompt(g) != nil {
			AnswerTriggerOrderInOfferedOrder(t, g)
			continue
		}
		if StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	var open []game.PendingChoiceKind
	for _, c := range g.PendingChoices {
		if c != nil {
			open = append(open, c.Kind)
		}
	}
	t.Fatalf("the stack did not settle: stack %d, pending triggers %d, open prompts %v", g.Stack.Size(), len(g.PendingTriggers), open)
}

// tctTokens lists the tokens controller controls.
func TctTokens(g *game.Game, controller uuid.UUID) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && effects.IsToken(c) {
				out = append(out, c)
			}
		}
	})
	return out
}
