package effects

// Dungeon Map — Artifact {3}:
//
//	"{T}: Add {C}.
//	 {3}, {T}: Venture into the dungeon. Activate only as a sorcery."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "20f1bba5-d398-4d6e-a76a-75e29df42aec",
		Name:         "Dungeon Map",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{3}, {T}: Venture into the dungeon. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}"), TapCost()),
			SorcerySpeed: true,
			Effect:       Venture,
		}},
	})
}
