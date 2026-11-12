package effects

// tiered_mana_lands.go — the Homelands land cycle (#5635): An-Havva
// Township, Aysen Abbey, Castle Sengir, Koskun Keep and Wizards' School
// all print
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add {<one>}.
//	 {2}, {T}: Add {<a>} or {<b>}."
//
// with a different colour and pair. Each row is an ordinary mana ability
// (CR 605.1a): the generic part of the cost is a ManaAbilityCost.Mana
// component paid from the pool before the land taps, and "{a} or {b}"
// is one pick, the pipe grammar Birds of Paradise uses.

// tieredManaLand is the three rows of one land in the cycle: `one` is
// the colour the {1} row adds, `a` and `b` the pair the {2} row picks
// between, each a single colour letter.
func tieredManaLand(one, a, b string) []ManaAbility {
	return []ManaAbility{
		colorlessTapRow(),
		{
			Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{" + one + "}",
			Label:    "{1}, {T}: Add {" + one + "}.",
		},
		{
			Cost:     ManaAbilityCost{Tap: true, Mana: "{2}"},
			Produced: "{" + a + "|" + b + "}",
			Label:    "{2}, {T}: Add {" + a + "} or {" + b + "}.",
		},
	}
}
