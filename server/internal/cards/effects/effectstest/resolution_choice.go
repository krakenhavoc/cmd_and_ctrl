package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_choice_test.go — #568. The option pick and the pile
// split, tested as shapes rather than only through the cards that use
// them: the chosen branch runs, the prompt is addressed to a seat that
// is not the controller, the split chains to a second prompt for a
// DIFFERENT seat, and both rewind.

// answerOptionPick answers the newest open option pick for `chooser`
// with the given index.
func AnswerOptionPick(t *testing.T, g *game.Game, chooser uuid.UUID, index int) {
	t.Helper()
	c := LatestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no option pick for %s: %+v", chooser, g.PendingChoices)
	}
	if err := g.ResolveOptionPick(c.ID, chooser, index); err != nil {
		t.Fatalf("ResolveOptionPick(%d): %v", index, err)
	}
}

// latestOptionPickFor returns the newest open option pick owed by a
// seat, or nil.
func LatestOptionPickFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	var out *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser == chooser {
			out = c
		}
	}
	return out
}

// latestChooseCardsFor returns the newest open choose-cards prompt
// owed by a seat, or nil.
func LatestChooseCardsFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	return LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, chooser)
}

// latestRevealPickFor is the same read for #1214's reveal_pick — "an
// opponent picks from a set you revealed", which is what a pile
// split's first leg has been since that kind existed.
func LatestRevealPickFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	return LatestChoiceOfKindFor(g, game.PendingChoiceRevealPick, chooser)
}

// latestChoiceOfKindFor is the body both of those share.
func LatestChoiceOfKindFor(g *game.Game, kind game.PendingChoiceKind, chooser uuid.UUID) *game.PendingChoice {
	var out *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind && c.Chooser == chooser {
			out = c
		}
	}
	return out
}
