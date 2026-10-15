package game

import (
	"testing"

	"github.com/google/uuid"
)

// graveyard_arrival_test.go — #3769, ADR 0066's 2026-10-15 amendment:
// a card remembers the zone it was put into a graveyard from, and that
// it happened this turn.

func arrivalCard() Card {
	return Card{InstanceID: uuid.New(), Name: "Wanderer", TypeLine: "Creature — Bear"}
}

func TestMoveCardStampsGraveyardArrival(t *testing.T) {
	for _, from := range []ZoneKind{ZoneHand, ZoneLibrary, ZoneBattlefield, ZoneStack, ZoneExile} {
		src, dst := newZone(from, uuid.Nil), newZone(ZoneGraveyard, uuid.Nil)
		c := arrivalCard()
		src.PushTop(c)
		moved, err := MoveCard(src, dst, c.InstanceID)
		if err != nil {
			t.Fatalf("%s -> graveyard: %v", from, err)
		}
		if moved.PutInGraveyardFrom != from || !moved.PutInGraveyardThisTurn {
			t.Errorf("%s -> graveyard: from %q, thisTurn %v", from, moved.PutInGraveyardFrom, moved.PutInGraveyardThisTurn)
		}
		if want := from != ZoneBattlefield; moved.PutInGraveyardNotFromBattlefieldThisTurn() != want {
			t.Errorf("%s -> graveyard: not-from-battlefield-this-turn = %v, want %v", from, !want, want)
		}
		if got := dst.Cards[0]; got.PutInGraveyardFrom != from {
			t.Errorf("%s: the card in the graveyard names %q", from, got.PutInGraveyardFrom)
		}
	}
}

// Leaving a graveyard forgets the record, so a card that comes back is a
// new object judged by its latest arrival only (CR 400.7).
func TestMoveCardForgetsGraveyardArrivalOnLeaving(t *testing.T) {
	grave, hand := newZone(ZoneGraveyard, uuid.Nil), newZone(ZoneHand, uuid.Nil)
	c := arrivalCard()
	c.PutInGraveyardFrom, c.PutInGraveyardThisTurn = ZoneBattlefield, true
	grave.PushTop(c)
	moved, err := MoveCard(grave, hand, c.InstanceID)
	if err != nil {
		t.Fatal(err)
	}
	if moved.PutInGraveyardFrom != "" || moved.PutInGraveyardThisTurn {
		t.Errorf("the record survived leaving the graveyard: %q %v", moved.PutInGraveyardFrom, moved.PutInGraveyardThisTurn)
	}
}

// A melded pair put into a graveyard stamps both of its cards.
func TestMeldedPairStampsBothCards(t *testing.T) {
	bf, grave := newZone(ZoneBattlefield, uuid.Nil), newZone(ZoneGraveyard, uuid.Nil)
	a, b := arrivalCard(), arrivalCard()
	m := Card{InstanceID: a.InstanceID, Name: "Melded", MeldedFrom: []Card{a, b}}
	bf.PushTop(m)
	if _, err := MoveCard(bf, grave, m.InstanceID); err != nil {
		t.Fatal(err)
	}
	if len(grave.Cards) != 2 {
		t.Fatalf("%d cards in the graveyard, want 2", len(grave.Cards))
	}
	for _, c := range grave.Cards {
		if c.PutInGraveyardFrom != ZoneBattlefield || !c.PutInGraveyardThisTurn {
			t.Errorf("%s: from %q thisTurn %v", c.InstanceID, c.PutInGraveyardFrom, c.PutInGraveyardThisTurn)
		}
	}
}

func TestPermissionFilterReadsTheArrivalRecord(t *testing.T) {
	f := PermissionFilter{PutInGraveyardThisTurn: true}
	cases := []struct {
		name string
		c    Card
		want bool
	}{
		{"from hand this turn", Card{PutInGraveyardFrom: ZoneHand, PutInGraveyardThisTurn: true}, true},
		{"from library this turn", Card{PutInGraveyardFrom: ZoneLibrary, PutInGraveyardThisTurn: true}, true},
		{"from the battlefield this turn", Card{PutInGraveyardFrom: ZoneBattlefield, PutInGraveyardThisTurn: true}, false},
		{"from hand, an earlier turn", Card{PutInGraveyardFrom: ZoneHand}, false},
		{"no record", Card{}, false},
	}
	for _, tc := range cases {
		if got := f.Matches(tc.c); got != tc.want {
			t.Errorf("%s: Matches = %v, want %v", tc.name, got, tc.want)
		}
		if !(PermissionFilter{}).Matches(tc.c) {
			t.Errorf("%s: the zero filter must match everything", tc.name)
		}
	}
}

// The "this turn" half ends as each turn begins, in every seat's
// graveyard; the source zone, a fact that does not expire, stays.
func TestGraveyardArrivalEndsWithTheTurn(t *testing.T) {
	g := newRestorableGame(t)
	var ids []uuid.UUID
	for _, p := range g.Seats {
		c := arrivalCard()
		c.PutInGraveyardFrom, c.PutInGraveyardThisTurn = ZoneHand, true
		p.Graveyard.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	// Walk the real turn machinery to the next turn rather than calling
	// the sweep directly, so dropping the hook from onTurnBeganLocked
	// fails this test.
	start := g.Turn.ActiveSeat
	began := g.Seats[start].TurnsBegun
	for i := 0; i < 64 && g.Turn.ActiveSeat == start; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat == start && g.Seats[start].TurnsBegun == began {
		t.Fatal("the harness did not reach another turn")
	}
	for i, p := range g.Seats {
		c := p.Graveyard.Cards[len(p.Graveyard.Cards)-1]
		if c.InstanceID != ids[i] {
			t.Fatalf("seat %d: unexpected card", i)
		}
		if c.PutInGraveyardThisTurn {
			t.Errorf("seat %d: still 'this turn' after the turn began", i)
		}
		if c.PutInGraveyardFrom != ZoneHand {
			t.Errorf("seat %d: the source zone was forgotten (%q)", i, c.PutInGraveyardFrom)
		}
	}
}

// A snapshot carries the record, so a game restored mid-turn still
// opens the cards Banon opens.
func TestGraveyardArrivalSurvivesSnapshotAndClone(t *testing.T) {
	g := newRestorableGame(t)
	p := g.Seats[0]
	c := arrivalCard()
	c.PutInGraveyardFrom, c.PutInGraveyardThisTurn = ZoneLibrary, true
	p.Graveyard.PushTop(c)

	_, restored := roundTrip(t, g)
	var got Card
	for _, rc := range restored.Seats[0].Graveyard.Cards {
		if rc.InstanceID == c.InstanceID {
			got = rc
		}
	}
	if got.PutInGraveyardFrom != ZoneLibrary || !got.PutInGraveyardThisTurn {
		t.Errorf("restored card: from %q thisTurn %v", got.PutInGraveyardFrom, got.PutInGraveyardThisTurn)
	}

	clone := g.Clone()
	for _, cc := range clone.Seats[0].Graveyard.Cards {
		if cc.InstanceID == c.InstanceID && (cc.PutInGraveyardFrom != ZoneLibrary || !cc.PutInGraveyardThisTurn) {
			t.Errorf("cloned card: from %q thisTurn %v", cc.PutInGraveyardFrom, cc.PutInGraveyardThisTurn)
		}
	}
}
