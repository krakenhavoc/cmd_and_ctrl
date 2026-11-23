package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The cost is the last shape: an among removal with NO printed kind.
func TestTekuthalDeclaresAnAnyKindAmongCounterCost(t *testing.T) {
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.TekuthalOracle})
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
// the counter makes it indestructible (CR 122.1b).
func TestTekuthalRemovesThreeCountersOfAnyKindsForIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tekuthal := effectstest.PushTekuthal(g, me.ID)
	bear := effectstest.B12Creature(g, me.ID, "Test Bear", "Creature — Bear", 2, 2)
	walker := effectstest.PushWalkerForCounterCost(g, me.ID, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.SetCounters(g, bear, map[string]int{game.CounterPlusOne: 2})

	effectstest.FloatForTest(g, me, "UUC")
	effectstest.B16Activate(t, g, me.ID, tekuthal, 0, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{bear, walker},
		CounterCounts:    []int{2, 1},
		CounterKinds:     []string{game.CounterPlusOne, game.CounterLoyalty},
		Strict:           true,
	})

	if effectstest.CounterCount(g, bear, game.CounterPlusOne) != 0 {
		t.Error("the Bear's +1/+1 counters did not pay")
	}
	if got := effectstest.CounterCount(g, walker, game.CounterLoyalty); got != 3 {
		t.Errorf("the walker has %d loyalty counters, want 3", got)
	}
	if got := effectstest.CounterCount(g, tekuthal, "indestructible"); got != 1 {
		t.Fatalf("Tekuthal has %d indestructible counters, want 1", got)
	}
	// CR 122.1b: the counter grants the keyword, read by the engine
	// itself (ADR 0101).
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, tekuthal), "indestructible") {
		t.Error("an indestructible counter did not make Tekuthal indestructible")
	}
}

// "Other": Tekuthal cannot pay with its own counters. The clause
// excludes the source by name, the posture every "another" cost
// clause in the catalog takes.
func TestTekuthalCannotPayWithItsOwnCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tekuthal := effectstest.PushTekuthal(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.SetCounters(g, tekuthal, map[string]int{game.CounterPlusOne: 5})

	effectstest.FloatForTest(g, me, "UUC")
	err := g.ActivateCatalogAbility(me.ID, tekuthal, 0, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{tekuthal},
		CounterCounts:    []int{3},
		CounterKinds:     []string{game.CounterPlusOne},
		Strict:           true,
	})
	if err == nil {
		t.Fatal("Tekuthal paid its own cost with its own counters")
	}
	if effectstest.CounterCount(g, tekuthal, game.CounterPlusOne) != 5 {
		t.Error("a refused payment removed counters")
	}
}

// CR 107.4f, #971: the {U/P} symbols may be paid with 2 life each on
// an ACTIVATION, so the ability fires off a blue-less pool.
func TestTekuthalPaysItsPhyrexianSymbolsWithLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tekuthal := effectstest.PushTekuthal(g, me.ID)
	bear := effectstest.B12Creature(g, me.ID, "Test Bear", "Creature — Bear", 2, 2)
	walker := effectstest.PushWalkerForCounterCost(g, me.ID, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.SetCounters(g, bear, map[string]int{game.CounterPlusOne: 2})

	life := me.Life
	effectstest.FloatForTest(g, me, "C") // the generic {1} only; both {U/P} go on life
	effectstest.B16Activate(t, g, me.ID, tekuthal, 0, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{bear, walker},
		CounterCounts:    []int{2, 1},
		CounterKinds:     []string{game.CounterPlusOne, game.CounterLoyalty},
		PhyrexianLife:    2,
		Strict:           true,
	})
	if got := life - me.Life; got != 2*game.PhyrexianLifePerSymbol {
		t.Errorf("life paid = %d, want %d for two Phyrexian symbols", got, 2*game.PhyrexianLifePerSymbol)
	}
	if effectstest.CounterCount(g, tekuthal, "indestructible") != 1 {
		t.Error("the ability did not resolve")
	}
}

// Nothing is declared any more. #943 shipped the card with one caveat
// — the proliferate replacement, which had no seam — and #976 built
// the seam, so every printed clause is modelled: the cost shape, the
// Phyrexian symbols (#971), the indestructible counter and the
// replacement.
func TestTekuthalDeclaresNoGap(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.TekuthalOracle)
	if !ok {
		t.Fatal("Tekuthal is not registered")
	}
	if spec.Completeness != effects.CompletenessFull {
		t.Errorf("completeness = %s, want full", spec.Completeness)
	}
	if len(spec.Caveats) != 0 {
		t.Errorf("caveats = %v, want none", spec.Caveats)
	}
	if len(spec.Replacements) != 1 {
		t.Fatalf("replacements = %d, want the one proliferate doubler", len(spec.Replacements))
	}
}
