package effects

// Woodland Mystic — Creature — Elf Druid {1}{G}, 1/1 (Core Set 2020, #5518):
//
//	"{T}: Add {G}."
//
// Llanowar Elves' mana ability: a single fixed {G}, so it adds the mana
// at once. Summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6996c354-d3fa-40c3-8534-27d2f5b74756",
		Name:         "Woodland Mystic",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
