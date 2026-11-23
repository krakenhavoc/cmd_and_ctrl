package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answerChooseCardsOrPick answers whichever sacrifice-style prompt is open
// with the given card.
func AnswerChooseCardsOrPick(t *testing.T, g *game.Game, chooser, id uuid.UUID) {
	t.Helper()
	if SacrificeChoiceFor(g, chooser) != nil {
		AnswerSacrifice(t, g, chooser, id)
		return
	}
	if pick := LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, chooser); pick != nil {
		if err := g.ResolveOwnPermanents(pick.ID, chooser, []uuid.UUID{id}); err != nil {
			t.Fatalf("ResolveOwnPermanents: %v", err)
		}
		return
	}
	if ChooseCardsChoiceFor(g, chooser) != nil {
		AnswerChooseCards(t, g, chooser, id)
		return
	}
	if LatestPickTarget(g, chooser) != nil {
		PickCard(t, g, chooser, id)
		return
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == chooser {
			t.Fatalf("unexpected open prompt %s", c.Kind)
		}
	}
	t.Fatal("no prompt is open")
}
