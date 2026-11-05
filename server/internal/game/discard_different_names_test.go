package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// discard_different_names_test.go — #4896: "Discard three cards with
// different names" (DiscardCost.DifferentNames). The set check is strict
// (CR 201.2b): every card needs a name, and no two may share one, with
// both halves of a split card counting (CR 201.2a, CR 709.4). The real
// card, Ormos, Archive Keeper, is proved end to end in
// cards/effects/ormos_archive_keeper_test.go.

func differentNamesClause(n int) *DiscardCost {
	return &DiscardCost{N: n, Label: "cards with different names", DifferentNames: true}
}

func namedHandCard(p *Player, name string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.Controller = p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}

func fireIceInHand(p *Player) uuid.UUID {
	c := splitFixture(p.ID,
		Face{Name: "Fire", TypeLine: "Instant", ManaCost: "{1}{R}"},
		Face{Name: "Ice", TypeLine: "Instant", ManaCost: "{1}{U}"})
	p.Hand.PushTop(c)
	return c.InstanceID
}

func TestDiscardNamesDifferIsStrict(t *testing.T) {
	split := splitFixture(uuid.New(),
		Face{Name: "Fire", TypeLine: "Instant", ManaCost: "{1}{R}"},
		Face{Name: "Ice", TypeLine: "Instant", ManaCost: "{1}{U}"})
	for _, tc := range []struct {
		name  string
		cards []Card
		want  bool
	}{
		{"three names", []Card{{Name: "Opt"}, {Name: "Shock"}, {Name: "Ponder"}}, true},
		{"a repeated name", []Card{{Name: "Opt"}, {Name: "Shock"}, {Name: "Opt"}}, false},
		{"a repeated name in another case", []Card{{Name: "Opt"}, {Name: "opt "}}, false},
		{"a card with no name (CR 201.2b)", []Card{{Name: "Opt"}, {Name: ""}}, false},
		{"a split card and one of its halves (CR 201.2a)", []Card{split, {Name: "Ice"}}, false},
		{"a split card and other names", []Card{split, {Name: "Opt"}, {Name: "Shock"}}, true},
	} {
		if got := DiscardNamesDiffer(tc.cards); got != tc.want {
			t.Errorf("%s: DiscardNamesDiffer = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestDiscardDifferentNamesValidatorRefusesARepeatedName(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	a := namedHandCard(me, "Opt")
	b := namedHandCard(me, "Opt")
	c := namedHandCard(me, "Shock")
	d := namedHandCard(me, "Ponder")
	cost := AbilityCost{DiscardCards: differentNamesClause(3)}
	src := uuid.New()

	var err error
	g.WithWriteLock(func() {
		_, err = g.validateDiscardCostLocked(me.ID, src, ZoneBattlefield, cost, []uuid.UUID{a, b, c}, 0)
	})
	if !errors.Is(err, ErrInvalidParam) || !errors.Is(err, errDiscardNamesRepeat) {
		t.Fatalf("two Opts: %v, want the different-names refusal", err)
	}
	var got []uuid.UUID
	g.WithWriteLock(func() {
		got, err = g.validateDiscardCostLocked(me.ID, src, ZoneBattlefield, cost, []uuid.UUID{a, c, d}, 0)
	})
	if err != nil || len(got) != 3 {
		t.Fatalf("three names: %v %v, want the three cards", got, err)
	}
}

func TestDiscardDifferentNamesValidatorReadsBothHalvesOfASplitCard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	split := fireIceInHand(me)
	fire := namedHandCard(me, "Fire")
	opt := namedHandCard(me, "Opt")
	cost := AbilityCost{DiscardCards: differentNamesClause(3)}

	var err error
	g.WithWriteLock(func() {
		_, err = g.validateDiscardCostLocked(me.ID, uuid.New(), ZoneBattlefield, cost, []uuid.UUID{split, fire, opt}, 0)
	})
	if !errors.Is(err, errDiscardNamesRepeat) {
		t.Fatalf("Fire // Ice beside Fire: %v, want the different-names refusal", err)
	}
}

func TestDiscardDifferentNamesOptionsDropANamelessCard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	opt := namedHandCard(me, "Opt")
	nameless := NewCard("", me.ID)
	me.Hand.PushTop(nameless)
	var got []uuid.UUID
	g.ReadSnapshot(func() { got = g.DiscardCostOptionsForEffect(me.ID, uuid.New(), differentNamesClause(2)) })
	if len(got) != 1 || got[0] != opt {
		t.Errorf("options = %v, want only the named card %v", got, opt)
	}
}

func TestDiscardDifferentNamesPaymentTakesTheFirstOfEachName(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	a := namedHandCard(me, "Opt")
	b := namedHandCard(me, "Opt")
	c := namedHandCard(me, "Shock")
	d := namedHandCard(me, "Shock")
	e := namedHandCard(me, "Ponder")
	cost := differentNamesClause(3)

	var got []uuid.UUID
	g.ReadSnapshot(func() { got = g.DiscardDifferentNamesPaymentForEffect(me.ID, cost, []uuid.UUID{a, b, c, d, e}) })
	if len(got) != 3 || got[0] != a || got[1] != c || got[2] != e {
		t.Errorf("payment = %v, want [%v %v %v]", got, a, c, e)
	}
	g.ReadSnapshot(func() { got = g.DiscardDifferentNamesPaymentForEffect(me.ID, cost, []uuid.UUID{a, b, c, d}) })
	if got != nil {
		t.Errorf("two names paid three: %v", got)
	}
}

// A split card first in the order would block the only set: the walk
// gives it up rather than report the cost unpayable.
func TestDiscardDifferentNamesPaymentBacktracksPastASplitCard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	split := fireIceInHand(me)
	fire := namedHandCard(me, "Fire")
	ice := namedHandCard(me, "Ice")

	var got []uuid.UUID
	g.ReadSnapshot(func() {
		got = g.DiscardDifferentNamesPaymentForEffect(me.ID, differentNamesClause(2), []uuid.UUID{split, fire, ice})
	})
	if len(got) != 2 || got[0] != fire || got[1] != ice {
		t.Errorf("payment = %v, want [Fire Ice]", got)
	}
}
