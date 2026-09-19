package effects

// Alloy Myr — 2/2 Artifact Creature — Myr for {3} (EDHREC rank 4018):
//
//	"{T}: Add one mana of any color."
//
// The colourless three-drop that fixes every colour, on a body. Five-
// colour decks play it over a Manalith because a creature can be
// tutored, blinked, sacrificed and pumped, and because it turns on the
// Myr payoffs.
//
// The pipe is the PRINTED width — all five colours — not a
// commander-identity narrowing. "Any color" and "any color in your
// commander's color identity" are different clauses and only the
// second one narrows (see ManaAbility.NarrowToCommanderIdentity); the
// client still lists the commander's colours first, which is a
// presentation order and not a restriction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "efb0394c-2a45-4dd8-bca3-08704056fa31",
		Name:         "Alloy Myr",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
