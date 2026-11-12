package effects

// Sisters of the Flame — Creature — Human Shaman {1}{R}{R}, 2/2
// (The Dark, #5638):
//
//	"{T}: Add {R}."
//
// Llanowar Elves' shape in red: one fixed colour, so the mana is added
// at once with no colour pick, and summoning sickness applies to the
// tap (Game.ActivateManaAbility checks it).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "389a9d46-d3fb-47f0-91cd-f6d487636916",
		Name:         "Sisters of the Flame",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
	})
}
