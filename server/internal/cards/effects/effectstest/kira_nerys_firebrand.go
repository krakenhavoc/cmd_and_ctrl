package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func EntryAttackPrompt(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceEntryAttackTarget {
			return c
		}
	}
	return nil
}

// putFromHand puts a card from its owner's hand onto the battlefield.
func PutFromHand(t *testing.T, g *game.Game, card, controller uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		_, err = g.PutFromHandOntoBattlefieldForEffect(card, game.HandEntryOptions{Controller: controller})
	})
	if err != nil {
		t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
	}
}

// answerEntryAttack picks the option naming `defender`.
func AnswerEntryAttack(t *testing.T, g *game.Game, defender uuid.UUID) {
	t.Helper()
	c := EntryAttackPrompt(g)
	if c == nil {
		t.Fatal("no entry_attack_target prompt is open")
	}
	for i, opt := range c.PickOptions {
		if opt.Player == defender || (len(opt.Cards) == 1 && opt.Cards[0] == defender) {
			if err := g.ResolveEntryAttackTarget(c.ID, c.Chooser, i); err != nil {
				t.Fatalf("ResolveEntryAttackTarget: %v", err)
			}
			return
		}
	}
	t.Fatalf("no option names %s: %+v", defender, c.PickOptions)
}
