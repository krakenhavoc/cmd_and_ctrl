package game

import "testing"

// TestPrintsManaText pins the per-card text scan behind
// Card.PrintsManaAbility (#2258).
func TestPrintsManaText(t *testing.T) {
	cases := []struct {
		name, typeLine, text string
		want                 bool
	}{
		{"mana rock", "Artifact", "{T}: Add {C}{C}.", true},
		{"any colour", "Artifact", "{T}: Add one mana of any color.", true},
		{"variable", "Creature — Elf", "{T}: Add X mana of any one color.", true},
		{"triggered", "Enchantment", "Whenever a land enters under your control, add {G}.", true},
		{"second line", "Artifact", "Flying\n{2}, {T}: Add {R}{R}.", true},
		{"reminder only", "Basic Land — Forest", "({T}: Add {G}.)", false},
		{"sorcery", "Sorcery", "Add {B}{B}{B}.", false},
		{"instant", "Instant", "Add {R}{R}.", false},
		{"no mana", "Creature — Bear", "When this enters, draw a card.", false},
		{"adds counters", "Creature — Bear", "Whenever you attack, add a +1/+1 counter.", false},
	}
	for _, tc := range cases {
		if got := PrintsManaText(tc.typeLine, tc.text); got != tc.want {
			t.Errorf("%s: PrintsManaText = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestMayMakeManaUnautomatedJoinsTheCatalog: the flag needs both the
// printed mana ability and a missing Spec (#2258).
func TestMayMakeManaUnautomatedJoinsTheCatalog(t *testing.T) {
	const registered = "9d2460c3-8eeb-4f35-b6f6-748c478664c7"
	prev := IsCatalogCard
	IsCatalogCard = func(oracleID string) bool { return oracleID == registered }
	t.Cleanup(func() { IsCatalogCard = prev })

	cases := []struct {
		name string
		card Card
		want bool
	}{
		{"uncatalogued rock", Card{OracleID: "x", NeedsEffect: true, PrintsManaAbility: true}, true},
		{"catalogued rock", Card{OracleID: registered, NeedsEffect: true, PrintsManaAbility: true}, false},
		{"uncatalogued, no mana", Card{OracleID: "x", NeedsEffect: true}, false},
		{"token", Card{Name: "Treasure"}, false},
	}
	for _, tc := range cases {
		if got := MayMakeManaUnautomated(tc.card); got != tc.want {
			t.Errorf("%s: = %v, want %v", tc.name, got, tc.want)
		}
	}
}
