package effects

// Gaea's Cradle — Legendary Land:
//
//	"{T}: Add {G} for each creature you control."
//
// Rank 446. The purest statement of the scaled shape — no cost beyond
// the tap, no gate, no restriction, just a count. Reserved-list, and
// the single most expensive land in green Commander, which is why it
// is worth having the engine get exactly right.
//
// With no creatures the land taps and adds nothing, which is the
// printed card and the reason ProducedFunc returning "" is a first-
// class answer rather than an error.
//
// The count is taken at activation, after the tap — so a creature
// that dies in response to something else is already gone, and one
// that entered a moment ago is already counted. CR 605.3a: a mana
// ability resolves the instant it is activated, so there is no window
// for the board to change between the count and the mana.
//
// Auto-tappable: the count is a pure read with no hidden cost and no
// decision, so the planner treats a Cradle as a green source of
// whatever size the board says.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "7c427c3d-ecd8-45ef-bebd-8f10f4a311db",
		Name:     "Gaea's Cradle",
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedPerPermanent("G", MatchCreature),
			Label:        "Add {G} for each creature you control",
		}},
	})
}
