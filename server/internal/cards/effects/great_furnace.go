package effects

// Great Furnace — Artifact Land (EDHREC rank 386):
//
//	"{T}: Add {R}."
//
// A Mountain that is also an artifact — which is the whole card:
// it counts for metalcraft (Dispatch), grows Storm-Kiln Artist, and
// dies to Vandalblast. The printed type line carries the artifact
// type; the only thing a Spec has to add is the mana ability the
// engine will not synthesise for a land without the basic
// supertype.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f4819061-b0b5-48ab-af7b-6525c3d2eab7",
		Name:         "Great Furnace",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
	})
}
