package effects

// Golden Hind — Creature — Elk {1}{G}, 2/1 (Journey into Nyx, #5558):
//
//	"{T}: Add {G}."
//
// Elvish Mystic's mana ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "82679258-345c-4062-ac62-e94a1e5ace78",
		Name:         "Golden Hind",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
