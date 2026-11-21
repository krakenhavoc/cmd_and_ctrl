package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// untap_static_cards_test.go exercises the printed cards which use the
// untap-step restriction seam.  untap_wave_test.go covers registration and
// one-shot effects; these tests keep the card-specific clauses observable.

func UntapStaticPendingPay(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		choice := g.PendingChoices[i]
		if choice != nil && choice.Kind == game.PendingChoicePayUnless && choice.Chooser == chooser {
			return choice
		}
	}
	t.Fatalf("no optional-payment prompt for %s", chooser)
	return nil
}

func UntapStaticTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("battlefield card %s is missing", id)
	}
	return c.Tapped
}
