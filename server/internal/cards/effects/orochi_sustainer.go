package effects

// Orochi Sustainer — Creature — Snake Shaman {1}{G}, 1/2 (Champions of
// Kamigawa, #5603):
//
//	"{T}: Add {G}."
//
// A mana ability (CR 605.1a) with {T}, so summoning sickness stops it
// (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d3039b42-25d5-4fe2-955a-0d910d7395bd",
		Name:         "Orochi Sustainer",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
