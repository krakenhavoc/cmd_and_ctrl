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
// DerivedColorsOnly exists as a separate flag on DerivedMatch rather
// than being baked into the derivation itself.
//
// A lone Reflecting Pool produces nothing: the only land you control
// is the Pool, and game.derivedManaLocked excludes a card from its own
// derivation. That is also the printed ruling, so this and the rules
// agree exactly, independently of the cycle guard below.
//
// Two Reflecting Pools, or a Pool and an Exotic Orchard the SAME
// player controls, DO now see each other when doing so is not a real
// cycle (#1323): the Pool's own-lands match reaches the Orchard, whose
// opponent-lands match then reads real opposing lands rather than
// looping back. Simplification, declared and narrower than it used to
// be: a genuinely CIRCULAR chain — two Pools facing each other, or a
// Pool and an Orchard that end up asking about each other across
// multiple opponents — still answers "no mana" for the permanents in
// the cycle (CR 106.6b), one colour short of what the real rules would
// resolve for some of those pairs. Weaker than printed.
func init() {
	Register(Spec{
		OracleID:     "67f43ac6-2a58-4b53-b5d7-0330e2a252e2",
		Name:         "Reflecting Pool",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Lands that copy other lands' mana in a genuine circle back to this Pool (facing another Reflecting Pool, or an Exotic Orchard that in turn reads this land) contribute nothing."},
		ManaAbilities: []ManaAbility{{
			Cost:              ManaAbilityCost{Tap: true},
			DerivedMatch:      DerivedFromOwnLands(),
			DerivedColorsOnly: false,
			// CR 106.6b: this ability reads what OTHER permanents
			// could produce, so CR 106.7's reader must route it
			// through the visited set rather than a plain ProducedFunc
			// call — which is also why a lone Pool makes nothing
			// (#782, #1323).
			DerivesFromOtherSources: true,
			Label:                   "Add one mana of any type a land you control could produce",
			// "any type that a land you control could produce" — no
			// mention of the commander's identity, so
			// NarrowToCommanderIdentity stays off.
		}},
	})
}
