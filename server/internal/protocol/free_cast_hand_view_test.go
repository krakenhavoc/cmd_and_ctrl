package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3270: a hand card picked for "cast a spell from your hand without paying
// its mana cost" carries the granted {0} offer on the wire, open in a step a
// sorcery could not be cast in (CR 608.2g), and the pick's prompt tells the
// client that answering it opens the cast.
func TestHandFreeCastOfferIsOpenWhereASorceryIsNot(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepEnd)
	if g.Seats[g.Turn.ActiveSeat].ID != me.ID {
		t.Skip("fixture: seat 0 is not the active seat")
	}

	c := game.NewCard("Test Sorcery", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{2}{R}"
	c.KnownBy = map[uuid.UUID]bool{me.ID: true}
	id := c.InstanceID
	me.Hand.PushTop(c)

	var granted bool
	g.WithWriteLock(func() {
		granted = g.FreeCastFromHandForEffect(me.ID, uuid.Nil, id, "Test Source")
	})
	if !granted {
		t.Fatal("the grant found no card in hand")
	}

	hand := ViewOfGameFor(g, me.ID.String()).Seats[0].Hand
	card := cardInSeatZone(t, hand, id)
	var offer *AlternativeCostView
	for i := range card.AlternativeCosts {
		if card.AlternativeCosts[i].Key == game.AltCostKeyGrantedCost {
			offer = &card.AlternativeCosts[i]
		}
	}
	if offer == nil {
		t.Fatalf("no granted-cost offer on the hand card: %v", keysOf(card.AlternativeCosts))
	}
	if offer.ManaCost != "{0}" {
		t.Errorf("the offer costs %q, want {0}", offer.ManaCost)
	}
	if offer.TimingClosed {
		t.Error("the free cast is closed in the end step; CR 608.2g ignores timing")
	}
	if !card.PrintedCostTimingClosed {
		t.Error("the printed sorcery cost must stay closed in the end step")
	}

	// The window closes on the holder's pass, and the card stays in hand.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
	card = cardInSeatZone(t, ViewOfGameFor(g, me.ID.String()).Seats[0].Hand, id)
	for _, o := range card.AlternativeCosts {
		if o.Key == game.AltCostKeyGrantedCost {
			t.Error("the granted offer survived the holder's pass")
		}
	}
}

func TestFreeCastPickPromptDeclaresTheCast(t *testing.T) {
	g := newTwoSeatGame(t)
	me, them := g.Seats[0], g.Seats[1]
	cards := []uuid.UUID{me.Hand.Cards[0].InstanceID}
	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:     me.ID,
			Question:    "cast a spell from your hand without paying its mana cost?",
			Cards:       cards,
			Max:         1,
			Zone:        game.ZoneHand,
			Destination: game.ChooseCastFree,
		})
	})
	if got := FilterViewFor(ViewOfGame(g), me.ID.String()).PendingChoices[0].ChooseDestination; got != "cast_free" {
		t.Errorf("chooser reads destination %q, want cast_free", got)
	}
	if got := FilterViewFor(ViewOfGame(g), them.ID.String()).PendingChoices[0].ChooseDestination; got != "" {
		t.Errorf("opponent reads destination %q, want it withheld", got)
	}
}
