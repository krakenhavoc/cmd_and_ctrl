package effects

// Opaline Unicorn — Artifact Creature — Unicorn {3}, 1/2 (Theros, #5561):
//
//	"{T}: Add one mana of any color."
//
// Alloy Myr's mana ability: one mana of any of the five colors.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca5ca2c9-2ddd-4e70-b738-8de61aa82570",
		Name:         "Opaline Unicorn",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
