package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// learn_test.go — the keyword action learn (CR 701.48a, #3525) on the
// outside-the-game seam, driven through Cram Session, whose only other
// clause is a life gain.

const (
	CramSessionOracle = "6e5b8bb2-9dd8-4986-9d0a-659114be5a2a"
	LessonTypeLine    = "Sorcery — Lesson"
)

// lessonCard is a Lesson in a sideboard.
func LessonCard(name string) game.Card { return OutsideCard(name, LessonTypeLine) }

// learnTable casts Cram Session for the active seat (seat 0) over a
// sideboard of `side`, and settles the spell. The life gain has
// happened, and the discard prompt is open when the hand has a card.
func LearnTable(t *testing.T, side ...game.Card) (*game.Game, *game.Player) {
	t.Helper()
	g := NewSideboardCatalogGame(t, 2, side...)
	me := g.Seats[0]
	life := me.Life
	CastCatalogSpell(t, g, "Cram Session", "Sorcery", CramSessionOracle, nil)
	PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("Cram Session's life gain: %d -> %d, want +4", life, me.Life)
	}
	return g, me
}

// learnDiscardPrompt is the open "you may discard a card" prompt: a
// choose_cards over the chooser's own hand.
func LearnDiscardPrompt(g *game.Game, me *game.Player) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID &&
			len(c.ChooseCards) > 0 && me.Hand.Contains(c.ChooseCards[0]) {
			return c
		}
	}
	return nil
}

// learnLessonPrompt is the open Lesson pick: a choose_cards whose
// candidates are not in the hand (they are outside the game).
func LearnLessonPrompt(g *game.Game, me *game.Player) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID &&
			len(c.ChooseCards) > 0 && !me.Hand.Contains(c.ChooseCards[0]) {
			return c
		}
	}
	return nil
}

// restoreRoundTripFrom is restoreRoundTrip for a snapshot taken earlier.
func RestoreRoundTripFrom(t *testing.T, snap *game.GameSnapshot) *game.Game {
	t.Helper()
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	restored.WithWriteLock(func() { restored.RecomputeLayersIfStaleLocked() })
	return restored
}
