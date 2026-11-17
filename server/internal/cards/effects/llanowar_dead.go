package effects

// Llanowar Dead — Creature — Zombie Elf {B}{G}, 2/2 (Apocalypse, #5614):
//
//	"{T}: Add {B}."
//
// Llanowar Elves' fixed one-colour mana ability, in black. Summoning
// sickness applies to the {T} cost (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a48541a-240b-4406-945c-1d0d032eca25",
		Name:         "Llanowar Dead",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
	})
}
