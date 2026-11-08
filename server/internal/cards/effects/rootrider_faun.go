package effects

// Rootrider Faun — Creature — Satyr Scout {1}{G}, 1/3 (Wilds of Eldraine,
// #4471):
//
//	"{T}: Add {G}.
//	 {1}, {T}: Add one mana of any color."
//
// Two mana abilities (CR 605.1a): the plain {G}, and Capital City's
// filter that turns one mana into one of any colour. Summoning sickness
// applies to both, since each has {T} in its cost (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "87bedc9b-b4d2-41b6-b9f1-e1aeb8e75ec3",
		Name:         "Rootrider Faun",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{G}",
				Label:    "Add {G}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
	})
}
