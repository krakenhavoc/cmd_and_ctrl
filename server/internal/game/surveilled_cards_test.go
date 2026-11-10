package game

import (
	"testing"

	"github.com/google/uuid"
)

// surveilled_cards_test.go — #5429: a card remembers that this turn's
// surveil put it into the graveyard (Card.Surveilled).

func surveilGraveyardCard(t *testing.T, p *Player, id uuid.UUID) Card {
	t.Helper()
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID == id {
			return c
		}
	}
	t.Fatalf("card %s is not in the graveyard", id)
	return Card{}
}

// The graveyard leg of a surveil marks the cards it bins. The card kept on
// top is not marked, and neither is a card milled from the same library.
func TestSurveilMarksTheCardsItPutsIntoTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	milled := libraryCard(me, "Milled")
	kept := libraryCard(me, "Kept")
	binned := libraryCard(me, "Binned")
	choice := openSurveil(t, g, me, 2, nil)
	if err := g.ResolveSurveil(choice.ID, me.ID, []uuid.UUID{binned}, []uuid.UUID{kept}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if c := surveilGraveyardCard(t, me, binned); !c.Surveilled || !c.SurveilledThisTurn() {
		t.Fatalf("the binned card: Surveilled %v, this turn %v", c.Surveilled, c.SurveilledThisTurn())
	}
	for _, c := range me.Library.Cards {
		if c.InstanceID == kept && c.Surveilled {
			t.Fatal("a card kept on top of the library was marked")
		}
	}
	g.WithWriteLock(func() {
		if err := g.MillNForEffect(me.ID, 2); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	if c := surveilGraveyardCard(t, me, kept); c.SurveilledThisTurn() {
		t.Fatal("a card that was looked at, kept, and then milled counts as surveilled into the graveyard")
	}
	if c := surveilGraveyardCard(t, me, milled); c.SurveilledThisTurn() {
		t.Fatal("a milled card counts as surveilled")
	}
}

// A leg a replacement exiles instead never reaches the graveyard and is
// not marked where it lands.
func TestSurveilDoesNotMarkACardAReplacementExiled(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	binned := libraryCard(me, "Binned")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(intoGraveyardReplacement(binned, "exile it instead", func(ev *ReplacementEvent) {
			ev.NewZone = ZoneExile
		}))
	})
	choice := openSurveil(t, g, me, 1, nil)
	if err := g.ResolveSurveil(choice.ID, me.ID, []uuid.UUID{binned}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if me.Graveyard.Contains(binned) {
		t.Fatal("non-vacuity: the replacement did not move the card elsewhere")
	}
	for _, c := range g.Exile.Cards {
		if c.InstanceID == binned && c.Surveilled {
			t.Fatal("a card exiled instead of surveilled into the graveyard was marked")
		}
	}
}

// The mark belongs to that arrival: a card that leaves the graveyard and
// comes back is a new object (CR 400.7).
func TestSurveilledMarkIsForgottenOnLeavingTheGraveyard(t *testing.T) {
	grave, hand := newZone(ZoneGraveyard, uuid.Nil), newZone(ZoneHand, uuid.Nil)
	c := arrivalCard()
	c.PutInGraveyardFrom, c.PutInGraveyardThisTurn, c.Surveilled = ZoneLibrary, true, true
	grave.PushTop(c)
	moved, err := MoveCard(grave, hand, c.InstanceID)
	if err != nil {
		t.Fatal(err)
	}
	if moved.Surveilled {
		t.Fatal("the mark survived leaving the graveyard")
	}
	back, err := MoveCard(hand, grave, c.InstanceID)
	if err != nil {
		t.Fatal(err)
	}
	if back.Surveilled || back.SurveilledThisTurn() {
		t.Fatal("a card discarded back into the graveyard counts as surveilled")
	}
}

func TestPermissionFilterReadsTheSurveilledMark(t *testing.T) {
	f := PermissionFilter{SurveilledThisTurn: true}
	for _, tc := range []struct {
		name string
		c    Card
		want bool
	}{
		{"surveilled this turn", Card{PutInGraveyardFrom: ZoneLibrary, PutInGraveyardThisTurn: true, Surveilled: true}, true},
		{"surveilled an earlier turn", Card{PutInGraveyardFrom: ZoneLibrary, Surveilled: true}, false},
		{"milled this turn", Card{PutInGraveyardFrom: ZoneLibrary, PutInGraveyardThisTurn: true}, false},
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

// "This turn" ends with the turn, and the mark is carried by a snapshot
// taken in between.
func TestSurveilledMarkSurvivesARestoreAndEndsWithTheTurn(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	binned := libraryCard(me, "Binned")
	choice := openSurveil(t, g, me, 1, nil)
	if err := g.ResolveSurveil(choice.ID, me.ID, []uuid.UUID{binned}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	_, restored := roundTrip(t, g)
	if c := surveilGraveyardCard(t, restored.Seats[restored.Turn.ActiveSeat], binned); !c.SurveilledThisTurn() {
		t.Fatal("the restored card is no longer surveilled this turn")
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
	if c := surveilGraveyardCard(t, me, binned); c.SurveilledThisTurn() {
		t.Fatal("the card is still surveilled this turn on the next turn")
	}
}
