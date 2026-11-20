package effects

// Leaf Gilder — Creature — Elf Druid {1}{G}, 2/1 (Lorwyn, #5593):
//
//	"{T}: Add {G}."
//
// Llanowar Elves' mana ability on a 2/1. As a creature's {T} ability it
// can't be used the turn the Gilder arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61324e37-4b79-4325-bf46-621b4270afe2",
		Name:         "Leaf Gilder",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
