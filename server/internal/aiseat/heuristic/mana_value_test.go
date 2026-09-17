package heuristic

import "testing"

// TestManaValueMatchesTheRules pins the policy-side mana-value parser
// to CR 202.3. The rows mirror internal/game's manaValueRows (this
// package may not import internal/game, ADR 0033 §3), plus the two
// shapes only this parser reads: a joined split-card cost and a
// hybrid Phyrexian symbol.
func TestManaValueMatchesTheRules(t *testing.T) {
	rows := []struct {
		name string
		cost string
		x    int
		want int
	}{
		{"no mana cost", "", 0, 0},
		{"{0}", "{0}", 0, 0},
		{"generic and coloured", "{3}{U}{U}", 0, 5},
		{"every colour", "{W}{U}{B}{R}{G}", 0, 5},
		{"colourless {C}", "{C}{C}", 0, 2},
		{"snow {S}", "{1}{S}{S}", 0, 3},
		{"{X} off the stack", "{X}{R}", 0, 1},
		{"{X} announced", "{X}{R}", 3, 4},
		{"{X}{X} announced", "{X}{X}{G}", 2, 5},
		{"two-colour hybrid is 1", "{1}{W/U}{W/U}", 0, 3},
		{"colourless hybrid {C/W} is 1", "{C/W}", 0, 1},
		{"monocoloured hybrid is 2", "{2/B}{2/B}{2/B}", 0, 6},
		{"monocoloured hybrid, lowercase", "{2/w}", 0, 2},
		{"Reaper King", "{2/W}{2/U}{2/B}{2/R}{2/G}", 0, 10},
		{"monocoloured hybrid beside generic", "{1}{2/U}{2/R}", 0, 5},
		{"Phyrexian is 1", "{1}{W/P}{W/P}", 0, 3},
		{"hybrid Phyrexian is 1", "{1}{G}{G/W/P}{W}", 0, 4},
		{"joined split cost sums its halves", "{1}{R} // {1}{U}", 0, 4},
	}
	for _, row := range rows {
		if got := manaValue(row.cost, row.x); got != row.want {
			t.Errorf("%s: manaValue(%q, %d) = %d, want %d", row.name, row.cost, row.x, got, row.want)
		}
	}
}
