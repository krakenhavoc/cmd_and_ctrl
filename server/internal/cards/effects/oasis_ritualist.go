package effects

// Oasis Ritualist — Creature — Snake Druid {3}{G}, 2/4:
//
//	"{T}: Add one mana of any color.
//	 {T}, Exert this creature: Add two mana of any one color. (An exerted
//	 creature won't untap during your next untap step.)"
//
// Both are mana abilities (CR 605.1a); the second pays an exert with
// the {T} (ADR 0130 §4, CR 701.43a). The auto-tapper plans only the
// first: an exert is a cost the player chooses, so the second is paid
// only when the player activates it from the creature's mana menu.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de89ef8f-ef6a-4f95-9c38-5debc06e1d74",
		Name:         "Oasis Ritualist",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}: Add one mana of any color.",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Exert: true},
				Produced: OneColorOfAmount(2),
				Label:    "{T}, Exert this creature: Add two mana of any one color.",
			},
		},
	})
}
