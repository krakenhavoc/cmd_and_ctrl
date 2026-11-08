package effects

// Three Tree Rootweaver — Creature — Mole Druid {1}{G}, 1/3 (Bloomburrow,
// #4467):
//
//	"{T}: Add one mana of any color."
//
// Birds of Paradise's mana ability on a body without flying. It is a
// creature's {T} ability, so summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d17d4003-e1c9-42be-ac76-3b5500fad485",
		Name:         "Three Tree Rootweaver",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
