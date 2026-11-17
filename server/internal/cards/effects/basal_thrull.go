package effects

// Basal Thrull — Creature — Thrull {B}{B}, 1/2 (Fallen Empires, #5637):
//
//	"{T}, Sacrifice this creature: Add {B}{B}."
//
// Crystal Vein's cash-in on a creature: a mana ability (CR 605.3b)
// whose cost taps and sacrifices the Thrull. The {T} means it can't be
// used the turn the Thrull arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9da50130-3f83-4968-983c-7dcba257cf1b",
		Name:         "Basal Thrull",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Sacrifice: true},
			Produced: "{B}{B}",
			Label:    "{T}, Sacrifice this creature: Add {B}{B}",
		}},
	})
}
