package game

import "testing"

// delve_test.go — the pure arithmetic behind delve (CR 702.66a, ADR
// 0100 §1). The board-level behaviour — validation, the exile at CR
// 601.2h, the record — is tested against real cards in
// cards/effects/delve_test.go.

func parseForDelveTest(t *testing.T, s string) ParsedCost {
	t.Helper()
	c, err := ParseCost(s)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", s, err)
	}
	return c
}

// Delve pays generic mana and nothing else: Treasure Cruise's {7}{U}
// has a budget of seven, and exiling seven leaves the {U}.
func TestDelveBudgetIsTheGenericOnly(t *testing.T) {
	cost := parseForDelveTest(t, "{7}{U}")
	if got := delveBudget(cost, 0); got != 7 {
		t.Fatalf("budget = %d, want 7", got)
	}
	left := delveAdjusted(cost, 7, 0)
	if left.Generic != 0 || len(left.Required) != 1 {
		t.Fatalf("after delving seven: %d generic + %v, want the {U} alone", left.Generic, left.Required)
	}
}

// X is folded in at the announced value, and {X}{X} is two slots:
// Empty the Pits at X = 3 owes six generic, all of it delvable.
func TestDelveBudgetFoldsX(t *testing.T) {
	cost := parseForDelveTest(t, "{X}{X}{B}{B}{B}{B}")
	if got := delveBudget(cost, 3); got != 6 {
		t.Fatalf("budget at X=3 = %d, want 6", got)
	}
	left := delveAdjusted(cost, 4, 3)
	if left.XSlots != 0 || left.Generic != 2 {
		t.Fatalf("after delving four at X=3: %d generic, %d X slots, want 2 and 0", left.Generic, left.XSlots)
	}
}

// Delving nothing leaves the cost exactly as it was, {X} slot and all,
// so a cast with no delve_ids prices as it did before delve existed.
func TestDelveNothingIsTheIdentity(t *testing.T) {
	cost := parseForDelveTest(t, "{X}{U}{U}")
	left := delveAdjusted(cost, 0, 5)
	if left.XSlots != 1 || left.Generic != 0 {
		t.Fatalf("delving nothing changed the cost: %+v", left)
	}
}

// Exiling more than the budget cannot drive the cost below zero; the
// announce validator refuses the over-delve outright, and this is the
// arithmetic's own floor.
func TestDelveCannotPayMoreThanTheGeneric(t *testing.T) {
	left := delveAdjusted(parseForDelveTest(t, "{2}{U}"), 5, 0)
	if left.Generic != 0 || len(left.Required) != 1 {
		t.Fatalf("over-delving produced %d generic + %v", left.Generic, left.Required)
	}
}

// A "spend mana as though it were mana of any colour" grant folds the
// coloured symbols into Generic for the solver. They are still
// coloured symbols in the rules, so delve may not pay them: a
// Breeches-granted Murktide Regent keeps its {U}{U}.
func TestDelveBudgetExcludesTheAnyColorFold(t *testing.T) {
	folded := asAnyColorCost(parseForDelveTest(t, "{5}{U}{U}"))
	if folded.Generic != 5 || len(folded.Required) != 2 {
		t.Fatalf("widening = %d generic, %d slots, want 5 and 2", folded.Generic, len(folded.Required))
	}
	if got := delveBudget(folded, 0); got != 5 {
		t.Fatalf("budget under the fold = %d, want 5", got)
	}
	typed := asAnyTypeCost(parseForDelveTest(t, "{3}{C}{G}"))
	if got := delveBudget(typed, 0); got != 3 {
		t.Fatalf("budget under the any-type fold = %d, want 3", got)
	}
}
