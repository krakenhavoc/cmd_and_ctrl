package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4483: Stern Scolding reads a creature SPELL's power and toughness, and
// a characteristic-defining ability works on the stack (CR 604.3:
// "function in all zones"; CR 113.6a, 208.2a). Tarmogoyf is `*/1+*`: its
// power is the number of card types among cards in all graveyards, its
// toughness that plus one. The importer's stand-in for the `*` is 0, so a
// predicate that reads CurrentPower on the stack sees 0/1 and wrongly
// allows a big Goyf.

const sternGoyfOracle = "45900b2f-f6a9-4c42-9642-008f3c1cf6dd"

// sternGoyfTypes is one graveyard card per card type, in the order the
// tests add them.
var sternGoyfTypes = []struct{ name, typeLine string }{
	{"Filler Artifact", "Artifact"},
	{"Filler Creature", "Creature — Bear"},
	{"Filler Land", "Basic Land — Forest"},
	{"Filler Instant", "Instant"},
	{"Filler Sorcery", "Sorcery"},
}

// sternFillGraveyard puts the types from index `from` up to `to` into seat
// 0's graveyard.
func sternFillGraveyard(g *game.Game, from, to int) {
	me := g.Seats[0]
	g.WithWriteLock(func() {
		for _, ct := range sternGoyfTypes[from:to] {
			pushGraveyardCardTyped(me, ct.name, ct.typeLine)
		}
	})
}

// sternCastGoyf casts a Tarmogoyf from seat 0's hand and leaves it on the
// stack.
func sternCastGoyf(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	toMainForCost(t, g)
	goyf := rfdHandCard(me, "Tarmogoyf", "Creature — Lhurgoyf", "", sternGoyfOracle, 0, 1)
	if err := g.CastSpell(me.ID, goyf, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Tarmogoyf: %v", err)
	}
	return goyf
}

func sternScold(t *testing.T, g *game.Game, spell uuid.UUID) error {
	t.Helper()
	_, err := hobB3Cast(t, g, g.Seats[0], "Stern Scolding", "Instant", "{U}", "{U}", game.CastSpellParams{Targets: b16TargetCard(spell)})
	return err
}

func TestHobSternScoldingReadsATarmogoyfSpellsCDA(t *testing.T) {
	for _, tc := range []struct {
		name  string
		types int // card types in graveyards: Goyf is types/(types+1)
		legal bool
	}{
		{"no types is 0/1", 0, true},
		{"two types is 2/3, power 2", 2, true},
		{"three types is 3/4", 3, false},
		{"five types is 5/6", 5, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			sternFillGraveyard(g, 0, tc.types)
			goyf := sternCastGoyf(t, g)
			err := sternScold(t, g, goyf)
			if tc.legal && err != nil {
				t.Fatalf("a Goyf spell at %d types is a legal target: %v", tc.types, err)
			}
			if !tc.legal && err == nil {
				t.Fatalf("Stern Scolding accepted a Goyf spell whose CDA makes it %d/%d", tc.types, tc.types+1)
			}
		})
	}
}

// ToughnessGE must read a creature spell's characteristic-defining toughness
// while the spell is on the stack.
func TestToughnessGEReadsATarmogoyfSpellCDA(t *testing.T) {
	g := newCatalogGame(t)
	sternFillGraveyard(g, 0, 3)
	sternCastGoyf(t, g)

	var spell game.Card
	g.WithWriteLock(func() {
		var err error
		spell, err = g.Stack.Top()
		if err != nil {
			t.Fatalf("get Tarmogoyf spell: %v", err)
		}
	})
	if spell.InstanceID == uuid.Nil {
		t.Fatal("expected the Tarmogoyf spell on the stack")
	}
	passes := func(n int) bool {
		var pass bool
		g.WithWriteLock(func() { pass = ToughnessGE(n)(g, uuid.Nil, spell) })
		return pass
	}
	if !passes(4) || passes(5) {
		t.Fatal("with three card types in graveyards, Tarmogoyf's 4 toughness should pass 4 and fail 5")
	}
	sternFillGraveyard(g, 3, 4)
	if !passes(5) {
		t.Fatal("with four card types in graveyards, Tarmogoyf's 5 toughness should pass 5")
	}
}

// The target is checked again on resolution: a Goyf that grows past 2/2
// while Stern Scolding waits is no longer a legal target, so the Scolding
// does nothing and the Goyf resolves.
func TestHobSternScoldingFizzlesWhenTheGoyfSpellGrows(t *testing.T) {
	g := newCatalogGame(t)
	sternFillGraveyard(g, 0, 2)
	goyf := sternCastGoyf(t, g)
	if err := sternScold(t, g, goyf); err != nil {
		t.Fatalf("a 2/3 Goyf spell is a legal target: %v", err)
	}
	sternFillGraveyard(g, 2, 3)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(goyf) {
		t.Error("the Goyf spell grew to 3/4, so Stern Scolding's target is illegal and the Goyf should resolve")
	}
}

// And the other way: a small Goyf spell is countered.
func TestHobSternScoldingCountersASmallTarmogoyfSpell(t *testing.T) {
	g := newCatalogGame(t)
	sternFillGraveyard(g, 0, 1)
	goyf := sternCastGoyf(t, g)
	if err := sternScold(t, g, goyf); err != nil {
		t.Fatalf("a 1/2 Goyf spell is a legal target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(goyf) || !g.Seats[0].Graveyard.Contains(goyf) {
		t.Error("the 1/2 Goyf spell should be countered")
	}
}
