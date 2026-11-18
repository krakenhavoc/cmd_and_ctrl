package effects

// Meteor Crater — Land:
//
//	"{T}: Choose a color of a permanent you control. Add one mana of that
//	 color."
//
// Mox Amber's shape: the offered colours are read off the board at
// activation, here the colours of every permanent its controller controls;
// the player picks one. With no coloured permanent it adds nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ebddf68c-1b22-4313-83e1-13a5c1aec19e",
		Name:         "Meteor Crater",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedFromOwnPermanentColors(),
			Label:        "Choose a color of a permanent you control. Add one mana of that color",
		}},
	})
}
