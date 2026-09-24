package effects

// Crystal Vein — Land (EDHREC rank 1510):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Add {C}{C}."
//
// A colourless land that can be cracked for a one-shot burst — the
// artifact deck's turn-three four-drop. Two mana abilities: the tap
// is Sol Ring's shape for one, the cash-in is Lotus Petal's
// sacrifice-the-source cost with two. Both resolve without the
// stack (CR 605.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "616d6013-24f4-4999-9bf3-5b0764e52fa6",
		Name:         "Crystal Vein",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Sacrifice: true},
				Produced: "{C}{C}",
				Label:    "{T}, Sacrifice this land: Add {C}{C}",
			},
		},
	})
}
