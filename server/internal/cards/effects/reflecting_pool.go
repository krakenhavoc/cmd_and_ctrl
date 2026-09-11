package effects

// Reflecting Pool — Land:
//
//	"{T}: Add one mana of any type that a land you control could
//	 produce."
//
// Rank 171. Exotic Orchard's question asked of your own board, with
// one word changed that changes the card: "any TYPE", not "any
// COLOR". Colorless is a type of mana but not a colour (CR 105.1), so
// a Reflecting Pool next to an Ancient Tomb can make {C} where an
// Exotic Orchard facing one could not. That single distinction is why
// mana_derivation.go keeps producibleFrom and colorsOnly as separate
// steps.
//
// A lone Reflecting Pool produces nothing: the only land you control
// is the Pool, its own ability is derived, and the recursion guard
// skips derived abilities. That is also the printed ruling, so here
// the guard and the rules agree exactly.
//
// Simplification, declared: two Reflecting Pools (or a Pool and an
// Exotic Orchard) do not see each other, for the same reason — the
// guard. The real rules resolve some of those pairs to a real colour;
// this is one colour short. Weaker than printed, and the alternative
// is an unbounded mutual recursion between two permanents.
func init() {
	Register(Spec{
		OracleID: "67f43ac6-2a58-4b53-b5d7-0330e2a252e2",
		Name:     "Reflecting Pool",
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedFromOwnLands(),
			Label:        "Add one mana of any type a land you control could produce",
			// "any type that a land you control could produce" — no
			// mention of the commander's identity, so no narrowing.
			IgnoreCommanderIdentity: true,
		}},
	})
}
