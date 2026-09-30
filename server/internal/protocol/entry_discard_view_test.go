package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entry_discard_view_test.go — ADR 0098's wire half. Mox Diamond's
// "discard a land card" prompt is over the chooser's own hand, so it
// takes the reveal prompt's redaction: the other seats see that a
// choice is open and whose it is, not the candidates or their count.

func TestEntryDiscardPromptIsPrivateToItsChooser(t *testing.T) {
	g := newTwoSeatGame(t)
	me, them := g.Seats[0], g.Seats[1]
	var cards []uuid.UUID
	for _, c := range me.Hand.Cards {
		if len(cards) == 2 {
			break
		}
		cards = append(cards, c.InstanceID)
	}
	if len(cards) != 2 {
		t.Fatalf("setup: wanted two cards in hand, found %d", len(cards))
	}
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind:        game.PendingChoiceEntryDiscardFromHand,
			Chooser:     me.ID,
			FromPlayer:  me.ID,
			Count:       1,
			Reason:      "Mox Diamond — discard a land card so it enters?",
			ChooseCards: cards,
			ChooseMin:   0,
			ChooseMax:   1,
		})
	})

	mine := FilterViewFor(ViewOfGame(g), me.ID.String())
	own := choiceFor(t, mine, string(game.PendingChoiceEntryDiscardFromHand))
	if len(own.Options) != len(cards) || own.ChooseMax != 1 {
		t.Errorf("chooser sees %d candidates, max %d; want %d, 1", len(own.Options), own.ChooseMax, len(cards))
	}

	theirs := FilterViewFor(ViewOfGame(g), them.ID.String())
	if len(theirs.PendingChoices) != 1 {
		t.Fatalf("opponent sees %d choices, want 1", len(theirs.PendingChoices))
	}
	other := theirs.PendingChoices[0]
	if len(other.Options) != 0 || other.ChooseMin != 0 || other.ChooseMax != 0 {
		t.Errorf("opponent reads %d candidates and bounds %d..%d; all of it is hidden-zone information",
			len(other.Options), other.ChooseMin, other.ChooseMax)
	}
}
