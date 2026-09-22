package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entry_reveal_view_test.go — #1198's wire half.
//
// "As this land enters, you may reveal an Island or Swamp card from
// your hand" asks about the revealer's OWN hand, and which of their
// cards match the clause is exactly the hidden information the
// question is about: an opponent who could read the candidate list —
// or merely its LENGTH — would know how many lands of two types the
// seat is holding before anything was revealed at all.
//
// So it takes choose_cards' redaction rather than untap_choice's
// (whose candidates are tapped permanents everybody can already see),
// even though it shares untap_choice's payload. What the other seats
// get is the same thing they get for a discard: that a choice is
// open, and whose it is. What was actually revealed reaches them
// afterwards, as the reveal's own grouped EventRevealCards run — the
// order CR 701.20 puts them in.

// queueEntryRevealView queues a prompt of the kind with `n`
// candidates out of the chooser's hand and returns their IDs.
func queueEntryRevealView(t *testing.T, g *game.Game, chooser *game.Player, n int) []uuid.UUID {
	t.Helper()
	var cards []uuid.UUID
	for _, c := range chooser.Hand.Cards {
		if len(cards) == n {
			break
		}
		cards = append(cards, c.InstanceID)
	}
	if len(cards) != n {
		t.Fatalf("setup: wanted %d cards in hand, found %d", n, len(cards))
	}
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind:        game.PendingChoiceEntryRevealFromHand,
			Chooser:     chooser.ID,
			FromPlayer:  chooser.ID,
			Count:       1,
			Reason:      "Choked Estuary — reveal an Island or Swamp card from your hand?",
			ChooseCards: cards,
			ChooseMin:   0,
			ChooseMax:   1,
		})
	})
	return cards
}

func TestEntryRevealPromptIsPrivateToItsChooser(t *testing.T) {
	g := newTwoSeatGame(t)
	me, them := g.Seats[0], g.Seats[1]
	cards := queueEntryRevealView(t, g, me, 3)

	mine := FilterViewFor(ViewOfGame(g), me.ID.String())
	own := choiceFor(t, mine, string(game.PendingChoiceEntryRevealFromHand))
	if len(own.Options) != len(cards) {
		t.Fatalf("chooser sees %d candidates, want %d", len(own.Options), len(cards))
	}
	for _, o := range own.Options {
		if o.Name == "" {
			t.Error("the chooser cannot read a card they are being asked to reveal")
		}
	}
	if own.ChooseMin != 0 || own.ChooseMax != 1 {
		t.Errorf("bounds reached the chooser as %d..%d, want 0..1", own.ChooseMin, own.ChooseMax)
	}
	if own.Reason == "" {
		t.Error("the question itself did not reach the wire")
	}

	theirs := FilterViewFor(ViewOfGame(g), them.ID.String())
	if len(theirs.PendingChoices) != 1 {
		t.Fatalf("opponent sees %d choices, want 1 (that a choice is open is public)",
			len(theirs.PendingChoices))
	}
	other := theirs.PendingChoices[0]
	if len(other.Options) != 0 {
		t.Errorf("opponent sees %d candidates; how many cards match the clause is hidden-zone information",
			len(other.Options))
	}
	if other.ChooseMin != 0 || other.ChooseMax != 0 {
		t.Errorf("opponent reads the bounds as %d..%d — that is the same leak by another name",
			other.ChooseMin, other.ChooseMax)
	}
	if other.Chooser != me.ID.String() {
		t.Error("the opponent cannot tell whose prompt is blocking the table")
	}
}
