package game

import "testing"

// x_cost_reduction_test.go — #2701. A generic cost reduction applies to
// the TOTAL cost, which counts {X} at its announced value:
//
//   - CR 107.3a: "While a spell is on the stack, any X in its mana cost
//     or in any alternative cost or additional cost it has equals the
//     announced value."
//   - CR 601.2f: "The total cost is the mana cost or alternative cost
//     (as determined in rule 601.2b), plus all additional costs and cost
//     increases, and minus all cost reductions."
//   - CR 118.7a: a generic reduction affects "only the generic mana
//     component of that cost" — never a coloured symbol.
//
// So the mana announced for X is generic mana a reduction may eat. X
// itself, the number on the stack, stays what was announced.

// reducerOnBoard wires a flat "spells cost {n} less" modifier onto a
// permanent of `me`'s.
func reducerOnBoard(t *testing.T, g *Game, me *Player, n int) {
	t.Helper()
	const oracle = "test-x-reducer"
	withCatalogCostModifiers(t, modifiersFor(oracle, CostModifier{
		Kind:   CostReduction,
		Label:  "Spells you cast cost less to cast.",
		Amount: fixed(n),
	}))
	modifierSource(t, g, me, "Test Reducer", oracle)
}

// payWith puts `generic` colourless and one of each colour in `colors`
// into the pool.
func payWith(me *Player, generic int, colors ...string) {
	me.ManaPool = nil
	for i := 0; i < generic; i++ {
		me.ManaPool.AddMana(ManaToken{Color: "C"})
	}
	for _, c := range colors {
		me.ManaPool.AddMana(ManaToken{Color: c})
	}
}

// The brief's case: {X}{G} under a {1} reduction at X = 0, 1 and 3.
// At X = 0 there is nothing generic to eat and the {G} stands; at X = 1
// the reduction pays the whole X; at X = 3 it pays one of three.
func TestAOneLessReductionComesOffX(t *testing.T) {
	for _, tc := range []struct {
		x, generic int
	}{{0, 0}, {1, 0}, {3, 2}} {
		g := newActiveGame(t)
		me := g.Seats[0]
		reducerOnBoard(t, g, me, 1)
		hydra := spellInHand(t, g, me, "Test Hydra", "Creature — Hydra", "{X}{G}")

		priced := priceOf(t, g, me, hydra, CastSpellParams{XValue: tc.x})
		if got := priced.GenericWithX(tc.x); got != tc.generic {
			t.Errorf("X = %d: generic owed %d, want %d", tc.x, got, tc.generic)
		}
		if len(priced.Required) != 1 {
			t.Errorf("X = %d: required %+v, want the {G} untouched", tc.x, priced.Required)
		}
		if priced.XSlots != 1 {
			t.Errorf("X = %d: the {X} slot is gone (%+v); the reduction must not settle X", tc.x, priced)
		}

		// One mana short of the price is refused under strict payment.
		if tc.generic > 0 {
			payWith(me, tc.generic-1, "G")
			if err := g.CastSpell(me.ID, hydra, CastSpellParams{Strict: true, XValue: tc.x}); err == nil {
				t.Fatalf("X = %d: cast with %d generic, want refused", tc.x, tc.generic-1)
			}
		}
		payWith(me, tc.generic, "G")
		if err := g.CastSpell(me.ID, hydra, CastSpellParams{Strict: true, XValue: tc.x}); err != nil {
			t.Fatalf("X = %d with {%d}{G}: %v", tc.x, tc.generic, err)
		}
		if n := len(me.ManaPool); n != 0 {
			t.Errorf("X = %d: %d mana left in the pool, want the exact price spent", tc.x, n)
		}
		if got := g.StackMeta[hydra].XValue; got != tc.x {
			t.Errorf("X on the stack = %d, want the announced %d", got, tc.x)
		}
	}
}

// The issue's case: {X}{G} announced at X = 3 costs {3}{G}, and a {2}
// reduction makes it {1}{G}.
func TestATwoLessReductionMakesX3Cost1G(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	reducerOnBoard(t, g, me, 2)
	hydra := spellInHand(t, g, me, "Test Hydra", "Creature — Hydra", "{X}{G}")
	priced := priceOf(t, g, me, hydra, CastSpellParams{XValue: 3})
	if got := priced.SettleX(3).String(); got != "{1}{G}" {
		t.Errorf("{X}{G} at X = 3 under {2} less = %s, want {1}{G}", got)
	}
}

// The printed generic goes first, then the X: {X}{1}{G} at X = 2 under
// {2} less is {1}{G}, one of it from the printed {1} and one from X.
// And {X}{X}{G} counts both X symbols (CR 107.3a): X = 2 is four
// generic, and {3} less leaves {1}{G}.
func TestAReductionEatsPrintedGenericThenX(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	reducerOnBoard(t, g, me, 2)
	a := spellInHand(t, g, me, "Test A", "Sorcery", "{X}{1}{G}")
	priced := priceOf(t, g, me, a, CastSpellParams{XValue: 2})
	if priced.Generic != 0 || priced.XReduced != 1 {
		t.Errorf("{X}{1}{G} under {2} less = %+v, want the {1} gone and one off X", priced)
	}
	if got := priced.SettleX(2).String(); got != "{1}{G}" {
		t.Errorf("{X}{1}{G} at X = 2 under {2} less = %s, want {1}{G}", got)
	}

	g2 := newActiveGame(t)
	me2 := g2.Seats[0]
	reducerOnBoard(t, g2, me2, 3)
	b := spellInHand(t, g2, me2, "Test B", "Sorcery", "{X}{X}{G}")
	if got := priceOf(t, g2, me2, b, CastSpellParams{XValue: 2}).SettleX(2).String(); got != "{1}{G}" {
		t.Errorf("{X}{X}{G} at X = 2 under {3} less = %s, want {1}{G}", got)
	}
}

// CR 118.7a and 601.2f: the reduction never reaches the coloured
// symbols and the cost never goes below {0}: {X}{G} at X = 1 under {5}
// less is {G}, and the overshoot is lost.
func TestAReductionOffXStopsAtTheColouredSymbols(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	reducerOnBoard(t, g, me, 5)
	hydra := spellInHand(t, g, me, "Test Hydra", "Creature — Hydra", "{X}{G}")
	priced := priceOf(t, g, me, hydra, CastSpellParams{XValue: 1})
	if got := priced.SettleX(1).String(); got != "{G}" {
		t.Errorf("{X}{G} at X = 1 under {5} less = %s, want {G}", got)
	}
	if got := priced.GenericWithX(1); got != 0 {
		t.Errorf("generic owed %d, want 0", got)
	}
}

// CR 601.2f order: increases, then reductions. A {1} tax on {X}{G} at
// X = 3 makes {4}{G}; {2} less takes the tax's {1} and one of X, so the
// cast pays {2}{G}.
func TestAnIncreaseThenAReductionOffX(t *testing.T) {
	const sphere, reducer = "test-sphere", "test-reducer"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCostModifiers(t, func(id string) []CostModifier {
		switch id {
		case sphere:
			return []CostModifier{{Kind: CostIncrease, Label: "+1", Amount: fixed(1)}}
		case reducer:
			return []CostModifier{{Kind: CostReduction, Label: "-2", Amount: fixed(2)}}
		}
		return nil
	})
	modifierSource(t, g, me, "Test Sphere", sphere)
	modifierSource(t, g, me, "Test Reducer", reducer)
	hydra := spellInHand(t, g, me, "Test Hydra", "Creature — Hydra", "{X}{G}")
	if got := priceOf(t, g, me, hydra, CastSpellParams{XValue: 3}).SettleX(3).String(); got != "{2}{G}" {
		t.Errorf("{X}{G} at X = 3 with +1 then -2 = %s, want {2}{G}", got)
	}
}

// Trinisphere measures the cost after the reduction off X: {X}{G} at
// X = 3 under {3} less is {G}, one mana, and the floor makes it {2}{G}.
func TestTrinisphereMeasuresTheCostAfterAReductionOffX(t *testing.T) {
	const trinisphere, reducer = "test-trinisphere", "test-reducer"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCostModifiers(t, func(id string) []CostModifier {
		switch id {
		case trinisphere:
			return []CostModifier{{Kind: CostFloor, Label: "at least three", Amount: fixed(3)}}
		case reducer:
			return []CostModifier{{Kind: CostReduction, Label: "-3", Amount: fixed(3)}}
		}
		return nil
	})
	modifierSource(t, g, me, "Test Trinisphere", trinisphere)
	modifierSource(t, g, me, "Test Reducer", reducer)
	hydra := spellInHand(t, g, me, "Test Hydra", "Creature — Hydra", "{X}{G}")
	if got := priceOf(t, g, me, hydra, CastSpellParams{XValue: 3}).SettleX(3).String(); got != "{2}{G}" {
		t.Errorf("{X}{G} at X = 3, {3} less, Trinisphere = %s, want {2}{G}", got)
	}
}

// The affordability probe and the payment read one number: with two
// mana and the reduction, {X}{G} is payable at X = 3 and not at X = 4.
func TestAffordabilityCountsAReductionOffX(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	reducerOnBoard(t, g, me, 2)
	hydra := spellInHand(t, g, me, "Test Hydra", "Creature — Hydra", "{X}{G}")
	payWith(me, 1, "G")
	for _, tc := range []struct {
		x    int
		want bool
	}{{3, true}, {4, false}} {
		priced := priceOf(t, g, me, hydra, CastSpellParams{XValue: tc.x})
		if got := me.ManaPool.CanPay(priced, tc.x); got != tc.want {
			t.Errorf("CanPay at X = %d: %v, want %v", tc.x, got, tc.want)
		}
	}
}

// "Spend only black mana on X" (spend_only.go): what the reduction took
// off the X is not paid, so it is not restricted either, and the rest
// of X stays black-only. {X} at X = 3 with one off X folds to two
// black-only symbols.
func TestSpendOnlyOnXFoldsWhatTheReductionLeft(t *testing.T) {
	c := ParsedCost{XSlots: 1, XReduced: 1, SpendOnly: &ManaSpendOnly{Colors: []string{"B"}, XOnly: true}}
	out := c.foldSpendOnly(3)
	if out.XSlots != 0 || out.XReduced != 0 || out.Generic != 0 {
		t.Errorf("folded %+v, want X settled", out)
	}
	if len(out.Required) != 2 {
		t.Fatalf("folded requirements %+v, want two black-only symbols", out.Required)
	}
	for _, r := range out.Required {
		if len(r.Options) != 1 || r.Options[0] != "B" {
			t.Errorf("requirement %+v, want black only", r)
		}
	}
}
