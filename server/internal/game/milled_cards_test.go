package game

import (
	"testing"

	"github.com/google/uuid"
)

// milled_cards_test.go — a card remembers that this turn's mill put it into
// the graveyard (Card.Milled, Raul, Trouble Shooter).

// A mill marks the cards it puts into the graveyard. A card put there by
// moving it out of a library any other way is not marked.
func TestMillMarksTheCardsItPutsIntoTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	milled := libraryCard(me, "Milled")
	g.WithWriteLock(func() {
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	if c := surveilGraveyardCard(t, me, milled); !c.Milled || !c.MilledThisTurn() {
		t.Fatalf("the milled card: Milled %v, this turn %v", c.Milled, c.MilledThisTurn())
	}
	// A surveil's graveyard leg is not a mill.
	binned := libraryCard(me, "Binned")
	choice := openSurveil(t, g, me, 1, nil)
	if err := g.ResolveSurveil(choice.ID, me.ID, []uuid.UUID{binned}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if c := surveilGraveyardCard(t, me, binned); c.MilledThisTurn() {
		t.Fatal("a surveilled card counts as milled")
	}
}

// A card exiled instead of milled never reaches the graveyard, so the mark is
// not on it where it lands.
func TestMillDoesNotMarkACardAReplacementExiled(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	milled := libraryCard(me, "Milled")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(intoGraveyardReplacement(milled, "exile it instead", func(ev *ReplacementEvent) {
			ev.NewZone = ZoneExile
		}))
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	if me.Graveyard.Contains(milled) {
		t.Fatal("non-vacuity: the replacement did not move the card elsewhere")
	}
	for _, c := range g.Exile.Cards {
		if c.InstanceID == milled && c.Milled {
			t.Fatal("a card exiled instead of milled was marked")
		}
	}
}

// The mark belongs to that arrival: a card that leaves the graveyard and comes
// back is a new object (CR 400.7).
func TestMilledMarkIsForgottenOnLeavingTheGraveyard(t *testing.T) {
	grave, hand := newZone(ZoneGraveyard, uuid.Nil), newZone(ZoneHand, uuid.Nil)
	c := arrivalCard()
	c.PutInGraveyardFrom, c.PutInGraveyardThisTurn, c.Milled = ZoneLibrary, true, true
	grave.PushTop(c)
	moved, err := MoveCard(grave, hand, c.InstanceID)
	if err != nil {
		t.Fatal(err)
	}
	if moved.Milled {
		t.Fatal("the mark survived leaving the graveyard")
	}
	back, err := MoveCard(hand, grave, c.InstanceID)
	if err != nil {
		t.Fatal(err)
	}
	if back.Milled || back.MilledThisTurn() {
		t.Fatal("a card discarded back into the graveyard counts as milled")
	}
}

func TestPermissionFilterReadsTheMilledMark(t *testing.T) {
	f := PermissionFilter{MilledThisTurn: true}
	for _, tc := range []struct {
		name string
		c    Card
		want bool
	}{
		{"milled this turn", Card{PutInGraveyardFrom: ZoneLibrary, PutInGraveyardThisTurn: true, Milled: true}, true},
		{"milled an earlier turn", Card{PutInGraveyardFrom: ZoneLibrary, Milled: true}, false},
		{"put there from a library some other way", Card{PutInGraveyardFrom: ZoneLibrary, PutInGraveyardThisTurn: true}, false},
		{"no record", Card{}, false},
	} {
		if got := f.Matches(tc.c); got != tc.want {
			t.Errorf("%s: Matches = %v, want %v", tc.name, got, tc.want)
		}
	}
	if !f.faceBlind() {
		t.Error("the clause is about the card where it sits, not its face")
	}
}

// "This turn" ends with the turn, and the mark is carried by a snapshot taken
// in between.
func TestMilledMarkSurvivesARestoreAndEndsWithTheTurn(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	milled := libraryCard(me, "Milled")
	g.WithWriteLock(func() {
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	_, restored := roundTrip(t, g)
	if c := surveilGraveyardCard(t, restored.Seats[restored.Turn.ActiveSeat], milled); !c.MilledThisTurn() {
		t.Fatal("the restored card is no longer milled this turn")
	}
	start := g.Turn.ActiveSeat
	for i := 0; i < 64 && g.Turn.ActiveSeat == start; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat == start {
		t.Fatal("the harness did not reach another turn")
	}
	if c := surveilGraveyardCard(t, me, milled); c.MilledThisTurn() {
		t.Fatal("the card is still milled this turn on the next turn")
	}
}
