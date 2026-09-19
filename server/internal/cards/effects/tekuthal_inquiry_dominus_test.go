package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tekuthal_inquiry_dominus_test.go — #943's card. The component's own
// rules are pinned in game/counter_cost_any_kind_test.go; this is
// about Tekuthal doing what its oracle text says.

const tekuthalOracle = "4716ab91-30e6-4c63-8389-a9db8f9414d8"

// pushTekuthal seats Tekuthal under `owner`'s control.
func pushTekuthal(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Tekuthal, Inquiry Dominus", "Legendary Creature — Phyrexian Horror", tekuthalOracle, 3, 5)
}

// The cost is the last shape: an among removal with NO printed kind.
func TestTekuthalDeclaresAnAnyKindAmongCounterCost(t *testing.T) {
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: tekuthalOracle})
	if len(abilities) != 1 {
		t.Fatalf("Tekuthal lists %d abilities, want one", len(abilities))
	}
	rc := abilities[0].Cost.RemoveCounters
	if rc == nil || !rc.Among || rc.Counter != "" || rc.N != 3 || rc.From == nil {
		t.Fatalf("cost = %+v, want three counters of any kind from among a clause", rc)
	}
	if abilities[0].Cost.Mana != "{1}{U/P}{U/P}" {
		t.Errorf("mana cost = %q, want {1}{U/P}{U/P}", abilities[0].Cost.Mana)
	}
	if _, err := game.ParseCost(abilities[0].Cost.Mana); err != nil {
		t.Errorf("the printed cost does not parse: %v", err)
	}
}

// The headline: three counters of two different kinds, off two
// different permanents, buy Tekuthal an indestructible counter — and
// the counter makes it indestructible (CR 122.1e).
func TestTekuthalRemovesThreeCountersOfAnyKindsForIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tekuthal := pushTekuthal(g, me.ID)
	bear := b12Creature(g, me.ID, "Test Bear", "Creature — Bear", 2, 2)
	walker := pushWalkerForCounterCost(g, me.ID, 4)
	advanceToMain(t, g)
	setCounters(g, bear, map[string]int{game.CounterPlusOne: 2})

	floatForTest(g, me, "UUC")
	b16Activate(t, g, me.ID, tekuthal, 0, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{bear, walker},
		CounterCounts:    []int{2, 1},
		CounterKinds:     []string{game.CounterPlusOne, game.CounterLoyalty},
		Strict:           true,
	})

	if counterCount(g, bear, game.CounterPlusOne) != 0 {
		t.Error("the Bear's +1/+1 counters did not pay")
	}
	if got := counterCount(g, walker, game.CounterLoyalty); got != 3 {
		t.Errorf("the walker has %d loyalty counters, want 3", got)
	}
	if got := counterCount(g, tekuthal, "indestructible"); got != 1 {
		t.Fatalf("Tekuthal has %d indestructible counters, want 1", got)
	}
	// CR 122.1e: the counter grants the keyword. The engine reads no
	// keyword counters of its own, so Tekuthal's own static carries
	// the rule (b24KeywordCounterGrant).
	if !containsString(effectiveAbilities(t, g, tekuthal), "indestructible") {
		t.Error("an indestructible counter did not make Tekuthal indestructible")
	}
}

// "Other": Tekuthal cannot pay with its own counters. The clause
// excludes the source by name, the posture every "another" cost
// clause in the catalog takes.
func TestTekuthalCannotPayWithItsOwnCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tekuthal := pushTekuthal(g, me.ID)
	advanceToMain(t, g)
	setCounters(g, tekuthal, map[string]int{game.CounterPlusOne: 5})

	floatForTest(g, me, "UUC")
	err := g.ActivateCatalogAbility(me.ID, tekuthal, 0, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{tekuthal},
		CounterCounts:    []int{3},
		CounterKinds:     []string{game.CounterPlusOne},
		Strict:           true,
	})
	if err == nil {
		t.Fatal("Tekuthal paid its own cost with its own counters")
	}
	if counterCount(g, tekuthal, game.CounterPlusOne) != 5 {
		t.Error("a refused payment removed counters")
	}
}

// CR 107.4f, #971: the {U/P} symbols may be paid with 2 life each on
// an ACTIVATION, so the ability fires off a blue-less pool.
func TestTekuthalPaysItsPhyrexianSymbolsWithLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tekuthal := pushTekuthal(g, me.ID)
	bear := b12Creature(g, me.ID, "Test Bear", "Creature — Bear", 2, 2)
	walker := pushWalkerForCounterCost(g, me.ID, 4)
	advanceToMain(t, g)
	setCounters(g, bear, map[string]int{game.CounterPlusOne: 2})

	life := me.Life
	floatForTest(g, me, "C") // the generic {1} only; both {U/P} go on life
	b16Activate(t, g, me.ID, tekuthal, 0, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{bear, walker},
		CounterCounts:    []int{2, 1},
		CounterKinds:     []string{game.CounterPlusOne, game.CounterLoyalty},
		PhyrexianLife:    2,
		Strict:           true,
	})
	if got := life - me.Life; got != 2*game.PhyrexianLifePerSymbol {
		t.Errorf("life paid = %d, want %d for two Phyrexian symbols", got, 2*game.PhyrexianLifePerSymbol)
	}
	if counterCount(g, tekuthal, "indestructible") != 1 {
		t.Error("the ability did not resolve")
	}
}

// The one declared gap is declared, and nothing else is: the card
// ships with a caveat naming the proliferate replacement, and nothing
// about the cost shape or the Phyrexian symbols.
func TestTekuthalDeclaresItsRemainingGap(t *testing.T) {
	spec, ok := Lookup(tekuthalOracle)
	if !ok {
		t.Fatal("Tekuthal is not registered")
	}
	if spec.Completeness != CompletenessCaveats {
		t.Errorf("completeness = %s, want caveats", spec.Completeness)
	}
	if len(spec.Caveats) != 1 {
		t.Fatalf("caveats = %v, want only the proliferate replacement", spec.Caveats)
	}
	if strings.Contains(spec.Caveats[0], "Phyrexian") {
		t.Errorf("the Phyrexian caveat outlived #971: %q", spec.Caveats[0])
	}
}
