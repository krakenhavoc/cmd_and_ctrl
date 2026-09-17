package game

import "testing"

// manaValueRows is CR 202.3 symbol by symbol, including the rule's
// own examples. aiseat/heuristic's TestManaValueMatchesTheRules
// carries a copy of the same rows for the policy-side parser, which
// may not import this package; keep the two in step.
var manaValueRows = []struct {
	name  string
	cost  string
	mv    int // off the stack: X is 0 (CR 202.3e)
	x     int // announced X for the on-stack read
	mvX   int // on the stack with X = x
	total int // what the cast charges at X = x (every hybrid paid with its coloured half)
}{
	{"no mana cost (CR 202.3a)", "", 0, 0, 0, 0},
	{"{0}", "{0}", 0, 0, 0, 0},
	{"generic and coloured (CR 202.3 example)", "{3}{U}{U}", 5, 0, 5, 5},
	{"every colour", "{W}{U}{B}{R}{G}", 5, 0, 5, 5},
	{"colourless {C}", "{C}{C}", 2, 0, 2, 2},
	{"snow {S}", "{1}{S}{S}", 3, 0, 3, 3},
	{"{X} is 0 off the stack, chosen value on it (CR 202.3e)", "{X}{R}", 1, 3, 4, 4},
	{"{X}{X}", "{X}{X}{G}", 1, 2, 5, 5},
	{"two-colour hybrid is 1 (CR 202.3f example)", "{1}{W/U}{W/U}", 3, 0, 3, 3},
	{"colourless hybrid {C/W} is 1", "{C/W}", 1, 0, 1, 1},
	{"monocoloured hybrid is 2 (CR 202.3f example)", "{2/B}{2/B}{2/B}", 6, 0, 6, 3},
	{"monocoloured hybrid, lowercase", "{2/w}", 2, 0, 2, 1},
	{"Reaper King", "{2/W}{2/U}{2/B}{2/R}{2/G}", 10, 0, 10, 5},
	{"monocoloured hybrid beside generic", "{1}{2/U}{2/R}", 5, 0, 5, 3},
	{"Phyrexian is 1 (CR 202.3g example)", "{1}{W/P}{W/P}", 3, 0, 3, 3},
}

func TestParsedCostManaValue(t *testing.T) {
	for _, row := range manaValueRows {
		t.Run(row.name, func(t *testing.T) {
			cost, err := ParseCost(row.cost)
			if err != nil {
				t.Fatalf("ParseCost(%q): %v", row.cost, err)
			}
			if got := cost.ManaValue(); got != row.mv {
				t.Errorf("ManaValue(%q) = %d, want %d", row.cost, got, row.mv)
			}
			if got := cost.ManaValueWithX(row.x); got != row.mvX {
				t.Errorf("ManaValueWithX(%q, %d) = %d, want %d", row.cost, row.x, got, row.mvX)
			}
			if got := cost.totalManaWithX(row.x); got != row.total {
				t.Errorf("totalManaWithX(%q, %d) = %d, want %d", row.cost, row.x, got, row.total)
			}

			card := Card{Name: row.name, ManaCost: row.cost}
			if got := card.ManaValue(); got != row.mv {
				t.Errorf("Card.ManaValue(%q) = %d, want %d", row.cost, got, row.mv)
			}
			if got, ok := card.ParsedManaValue(); !ok || got != row.mv {
				t.Errorf("Card.ParsedManaValue(%q) = %d, %v; want %d, true", row.cost, got, ok, row.mv)
			}
			if got := card.ManaValueWithX(row.x); got != row.mvX {
				t.Errorf("Card.ManaValueWithX(%q, %d) = %d, want %d", row.cost, row.x, got, row.mvX)
			}
		})
	}
}

// TestCardManaValueOfAnUnreadableCost pins the two readings of a cost
// the parser rejects (a joined split-card cost): ManaValue reports
// zero, ParsedManaValue reports not-ok so a "mana value N or less"
// predicate can refuse the card rather than pass it as zero.
func TestCardManaValueOfAnUnreadableCost(t *testing.T) {
	card := Card{Name: "Fire // Ice", ManaCost: "{1}{R} // {1}{U}"}
	if got := card.ManaValue(); got != 0 {
		t.Errorf("ManaValue = %d, want 0", got)
	}
	if got, ok := card.ParsedManaValue(); ok || got != 0 {
		t.Errorf("ParsedManaValue = %d, %v; want 0, false", got, ok)
	}
	if got := card.ManaValueWithX(3); got != 0 {
		t.Errorf("ManaValueWithX = %d, want 0", got)
	}
}

// TestCostFloorMeasuresTheTotalCostNotTheManaValue: Trinisphere reads
// what the cast charges (CR 601.2f), and the engine charges a {2/W}
// as its coloured half. A lone {2/W} is mana value 2 but costs one,
// so a three-floor adds two generic — not the one a mana-value read
// would add.
func TestCostFloorMeasuresTheTotalCostNotTheManaValue(t *testing.T) {
	cost, err := ParseCost("{2/W}")
	if err != nil {
		t.Fatal(err)
	}
	got := raiseToMinimum(cost, 3, 0)
	if got.Generic != 2 || len(got.Required) != 1 {
		t.Errorf("{2/W} under a three-floor = generic %d + %d coloured, want generic 2 + 1 coloured",
			got.Generic, len(got.Required))
	}

	// Spectral Procession already costs three paid with its coloured
	// halves; the floor adds nothing.
	cost, err = ParseCost("{2/W}{2/W}{2/W}")
	if err != nil {
		t.Fatal(err)
	}
	if got := raiseToMinimum(cost, 3, 0); got.Generic != 0 {
		t.Errorf("{2/W}{2/W}{2/W} under a three-floor gained %d generic, want 0", got.Generic)
	}
}
