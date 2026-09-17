package effects

// Tarnished Citadel — Land (EDHREC rank 2349):
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. This land deals 3 damage to you."
//
// City of Brass with a painless colourless half and a three-point
// rider on the coloured one. The rider is the painland shape
// (PainRider) at 3: damage that happens whether or not the player
// could "afford" it, from the Citadel itself, so a damage doubler
// sees it. "Any color" is the printed width, so the pipe is not
// narrowed to the commander's identity.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "66ae2562-68e9-4c77-ba0a-57f8ff37f656",
		Name:         "Tarnished Citadel",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color. This land deals 3 damage to you",
				Rider:    PainRider(3),
			},
		},
	})
}
