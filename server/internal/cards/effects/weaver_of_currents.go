package effects

// Weaver of Currents — Creature — Snake Druid {1}{G}{U}, 2/2
// (Amonkhet, #5535):
//
//	"{T}: Add {C}{C}."
//
// A tap mana ability, the shape of Worn Powerstone's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9aabb520-7e97-47b2-975c-11f457d0b870",
		Name:         "Weaver of Currents",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
	})
}
