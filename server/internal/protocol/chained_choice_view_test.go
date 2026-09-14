package protocol

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func newTwoSeatGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := make([]game.Card, 10)
		for j := range deck {
			deck[j] = game.NewCard("filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer %d: %v", i, err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(5, 6))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

// TestConfirmPromptCarriesItsBranchLabels — the card's own words reach
// the client, because "Pay 4 life" / "Put it on top" is the question
// and "Yes" / "No" is not.
func TestConfirmPromptCarriesItsBranchLabels(t *testing.T) {
	g := newTwoSeatGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueConfirmForEffect(game.ConfirmPrompt{
			Chooser:      me.ID,
			Question:     "Sylvan Library — pay 4 life to keep it?",
			AcceptLabel:  "Pay 4 life",
			DeclineLabel: "Put it on top",
		})
	})
	v := FilterViewFor(ViewOfGame(g), me.ID.String())
	if len(v.PendingChoices) != 1 {
		t.Fatalf("saw %d choices, want 1", len(v.PendingChoices))
	}
	c := v.PendingChoices[0]
	if c.Kind != "confirm" {
		t.Fatalf("kind = %q, want confirm", c.Kind)
	}
	if c.AcceptLabel != "Pay 4 life" || c.DeclineLabel != "Put it on top" {
		t.Errorf("labels = %q / %q", c.AcceptLabel, c.DeclineLabel)
	}
	if c.Reason == "" {
		t.Error("the question itself did not reach the wire")
	}
}

// TestChooseCardsPromptIsPrivateToItsChooser — the candidates are cards
// in a hand, so an opponent must learn neither what they are nor how
// many there are. Redaction alone is not enough: a list of backs still
// has a length, and the bounds would say "2 of 3" about a hidden zone.
func TestChooseCardsPromptIsPrivateToItsChooser(t *testing.T) {
	g := newTwoSeatGame(t)
	me, them := g.Seats[0], g.Seats[1]
	var cards []uuid.UUID
	for _, c := range me.Hand.Cards {
		if len(cards) == 3 {
			break
		}
		cards = append(cards, c.InstanceID)
	}
	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  me.ID,
			Question: "choose two",
			Cards:    cards,
			Min:      2,
			Max:      2,
			Zone:     game.ZoneHand,
		})
	})

	mine := FilterViewFor(ViewOfGame(g), me.ID.String())
	own := mine.PendingChoices[0]
	if len(own.Options) != len(cards) {
		t.Fatalf("chooser sees %d candidates, want %d", len(own.Options), len(cards))
	}
	for _, o := range own.Options {
		if o.Name == "" {
			t.Error("the chooser cannot read a card they are being asked to choose")
		}
	}
	if own.ChooseMin != 2 || own.ChooseMax != 2 {
		t.Errorf("bounds reached the chooser as %d..%d", own.ChooseMin, own.ChooseMax)
	}

	theirs := FilterViewFor(ViewOfGame(g), them.ID.String())
	if len(theirs.PendingChoices) != 1 {
		t.Fatalf("opponent sees %d choices, want 1 (that a choice is open is public)",
			len(theirs.PendingChoices))
	}
	other := theirs.PendingChoices[0]
	if len(other.Options) != 0 {
		t.Errorf("opponent sees %d candidates; the COUNT alone is hidden-zone information",
			len(other.Options))
	}
	if other.ChooseMin != 0 || other.ChooseMax != 0 {
		t.Errorf("opponent reads the bounds as %d..%d — that is the same leak by another name",
			other.ChooseMin, other.ChooseMax)
	}
}
