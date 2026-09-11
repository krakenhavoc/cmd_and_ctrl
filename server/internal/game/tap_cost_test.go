package game

import "testing"

// tap_cost_test.go — the pure arithmetic behind convoke and
// waterbend: what a tapped permanent buys, and which part of the
// cost it is allowed to buy it from. The board-level behaviour
// (validation, tapping, the strict gate) is tested against real
// cards in cards/effects/tap_cost_test.go.

func convokeForTest() *TapPermanentsCost {
	return &TapPermanentsCost{
		Key:         "convoke",
		ColorClause: true,
		Spec:        &TargetSpec{Zones: []ZoneKind{ZoneBattlefield}},
	}
}

func waterbendForTest(extra string) *TapPermanentsCost {
	return &TapPermanentsCost{
		Key:   "waterbend",
		Extra: extra,
		Spec:  &TargetSpec{Zones: []ZoneKind{ZoneBattlefield}},
	}
}

func parseCostForTapTest(t *testing.T, s string) ParsedCost {
	t.Helper()
	cost, err := ParseCost(s)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", s, err)
	}
	return cost
}

// Convoke's colour clause is the whole reason the assignment is a
// matching problem: {W}{U} paid by a white-blue creature and a white
// creature is payable in full, but only if the white-blue one is put
// on the {U}. A greedy first-match walk hands it to the {W} and
// convokes one creature instead of two.
func TestConvokeAssignsColorsByMaximumMatching(t *testing.T) {
	payers := []Card{
		{Colors: []string{"W", "U"}},
		{Colors: []string{"W"}},
	}
	got := tapPermanentsAdjusted(parseCostForTapTest(t, "{W}{U}"), convokeForTest(), payers, 0)
	if got.Generic != 0 || len(got.Required) != 0 {
		t.Fatalf("two creatures at {W}{U} left %d generic + %v colored, want nothing owed",
			got.Generic, got.Required)
	}
}

// A colourless creature has no colour, so it can pay {1} and nothing
// else. {1}{W} convoked with one colourless creature still owes the
// {W}.
func TestConvokeColorlessCreaturePaysOnlyGeneric(t *testing.T) {
	got := tapPermanentsAdjusted(parseCostForTapTest(t, "{1}{W}"), convokeForTest(), []Card{{}}, 0)
	if got.Generic != 0 {
		t.Errorf("generic left %d, want 0", got.Generic)
	}
	if len(got.Required) != 1 || got.Required[0].Options[0] != "W" {
		t.Errorf("colored left %v, want the {W} still owed", got.Required)
	}
}

// A hybrid symbol is covered by either half, and a single creature
// of either colour pays it.
func TestConvokeCoversAHybridSymbolWithEitherColor(t *testing.T) {
	for _, color := range []string{"W", "U"} {
		got := tapPermanentsAdjusted(parseCostForTapTest(t, "{W/U}"), convokeForTest(),
			[]Card{{Colors: []string{color}}}, 0)
		if len(got.Required) != 0 {
			t.Errorf("a %s creature did not cover {W/U}: %v", color, got.Required)
		}
	}
}

// Tapping more creatures than the cost has symbols cannot drive the
// cost below zero. (The announce validator rejects the over-tap
// outright; this is the arithmetic's own floor.)
func TestConvokeCannotReduceBelowZero(t *testing.T) {
	payers := []Card{{Colors: []string{"W"}}, {Colors: []string{"W"}}, {Colors: []string{"W"}}}
	got := tapPermanentsAdjusted(parseCostForTapTest(t, "{W}"), convokeForTest(), payers, 0)
	if got.Generic != 0 || len(got.Required) != 0 {
		t.Fatalf("over-tapping produced %d generic + %v colored, want nothing owed",
			got.Generic, got.Required)
	}
}

// Waterbend has no colour clause, and — the part that matters — it
// pays its OWN cost, never the card's. Waterbender's Restoration is
// {U}{U} plus waterbend {X}: tapping three permanents at X=3 clears
// the waterbend and leaves the {U}{U} exactly where it was.
func TestWaterbendPaysItsOwnCostAndNotTheSpells(t *testing.T) {
	payers := []Card{{Colors: []string{"U"}}, {Colors: []string{"U"}}, {}}
	got := tapPermanentsAdjusted(parseCostForTapTest(t, "{U}{U}"), waterbendForTest("{X}"), payers, 3)
	if got.Generic != 0 {
		t.Errorf("waterbend {3} paid by three permanents left %d generic, want 0", got.Generic)
	}
	if len(got.Required) != 2 {
		t.Errorf("the spell's own {U}{U} was discounted: %v", got.Required)
	}
}

// Tapping fewer permanents than X leaves the remainder owed in mana.
func TestWaterbendRemainderIsStillOwedInMana(t *testing.T) {
	got := tapPermanentsAdjusted(parseCostForTapTest(t, "{U}{U}"), waterbendForTest("{X}"), []Card{{}}, 3)
	if got.Generic != 2 {
		t.Errorf("waterbend {3} with one permanent tapped owes %d generic, want 2", got.Generic)
	}
}

// Tapping nothing leaves the waterbend cost owed in full — the cast
// is still legal, it is just paid entirely from the pool.
func TestWaterbendWithNoTapsOwesTheWholeExtra(t *testing.T) {
	got := tapPermanentsAdjusted(parseCostForTapTest(t, "{U}{U}"), waterbendForTest("{X}"), nil, 4)
	if got.Generic != 4 {
		t.Errorf("untapped waterbend {4} owes %d generic, want 4", got.Generic)
	}
}

// The budget is what caps the picker and the validator: the mana
// value of the spell for convoke, the size of the keyword's own cost
// for waterbend.
func TestTapPermanentsBudget(t *testing.T) {
	if got := TapPermanentsBudgetFor(convokeForTest(), "{3}{W}{W}", 0); got != 5 {
		t.Errorf("convoke budget at {3}{W}{W} = %d, want 5", got)
	}
	if got := TapPermanentsBudgetFor(waterbendForTest("{X}"), "{U}{U}", 3); got != 3 {
		t.Errorf("waterbend {X} budget at X=3 = %d, want 3", got)
	}
	// A waterbend cost's budget does not grow with the card's own
	// cost — tapping can only ever pay the keyword's half.
	if got := TapPermanentsBudgetFor(waterbendForTest("{4}"), "{6}{U}{U}", 0); got != 4 {
		t.Errorf("waterbend {4} budget = %d, want 4", got)
	}
	if got := TapPermanentsBudgetFor(nil, "{3}{W}{W}", 0); got != 0 {
		t.Errorf("nil cost budget = %d, want 0", got)
	}
}

// DemandsX is what tells the client to open the X prompt for a card
// whose PRINTED cost has no {X} in it.
func TestTapCostDemandsX(t *testing.T) {
	if !waterbendForTest("{X}").DemandsX() {
		t.Error("waterbend {X} does not demand an X")
	}
	if waterbendForTest("{4}").DemandsX() {
		t.Error("waterbend {4} demands an X it does not have")
	}
	if convokeForTest().DemandsX() {
		t.Error("convoke demands an X")
	}
	var nilCost *TapPermanentsCost
	if nilCost.DemandsX() || !nilCost.Empty() {
		t.Error("the nil cost is not inert")
	}
}
