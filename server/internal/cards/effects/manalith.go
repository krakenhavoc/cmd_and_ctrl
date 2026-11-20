package effects

// Manalith — Artifact {3} (Magic 2012, #5573):
//
//	"{T}: Add one mana of any color."
//
// Alloy Myr's mana ability on an artifact: {T} for one mana of any color.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bd9e416a-89b3-4912-be9b-49fce6a93dc9",
		Name:         "Manalith",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
