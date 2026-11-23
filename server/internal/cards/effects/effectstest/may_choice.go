package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// may_choice_test.go — #796. The free yes/no a resolving effect asks,
// tested as a primitive rather than only through the three cards that
// use it: the branches run once and in order, the prompt blocks the
// table, the question can be addressed to a seat other than the
// controller, and the whole thing rewinds.

// answerMayChoice answers the newest open MayChoice (a
// PendingChoiceConfirm) for `chooser`.
func AnswerMayChoice(t *testing.T, g *game.Game, chooser uuid.UUID, yes bool) {
	t.Helper()
	c := LatestConfirmFor(g, chooser)
	if c == nil {
		t.Fatalf("no yes/no prompt for %s: %+v", chooser, g.PendingChoices)
	}
	if err := g.ResolveConfirm(c.ID, chooser, yes); err != nil {
		t.Fatalf("ResolveConfirm(%v): %v", yes, err)
	}
}

// latestConfirmFor returns the newest open confirm prompt owed by a
// seat, or nil.
func LatestConfirmFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	var out *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm && c.Chooser == chooser {
			out = c
		}
	}
	return out
}
