package game

import "testing"

// s82_gap_b154_test.go — the two small seams GAP batch 154 (#5948) adds for
// Isu the Abominable: a three-way hybrid cost symbol, and a snow filter on a
// cast permission.

func TestParseCostReadsAThreeWayHybridSymbol(t *testing.T) {
	cost, err := ParseCost("{G/W/U}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	if len(cost.Required) != 1 || len(cost.Required[0].Options) != 3 {
		t.Fatalf("required = %+v, want one slot with three options", cost.Required)
	}
	if got := cost.ManaValueWithX(0); got != 1 {
		t.Errorf("mana value %d, want 1", got)
	}
	for _, color := range []string{"G", "W", "U"} {
		if !cost.Required[0].Admits(color) {
			t.Errorf("{G/W/U} does not admit %s", color)
		}
	}
	if cost.Required[0].Admits("R") {
		t.Error("{G/W/U} admits R")
	}
}

func TestPermissionFilterSnowOnlyJudgesTheSupertype(t *testing.T) {
	f := PermissionFilter{SnowOnly: true}
	if !f.Matches(Card{Name: "Snow-Covered Forest", TypeLine: "Basic Snow Land — Forest"}) {
		t.Error("a snow land does not match")
	}
	if !f.Matches(Card{Name: "Icy Bear", TypeLine: "Snow Creature — Bear"}) {
		t.Error("a snow creature does not match")
	}
	if f.Matches(Card{Name: "Forest", TypeLine: "Basic Land — Forest"}) {
		t.Error("a nonsnow land matches")
	}
}
