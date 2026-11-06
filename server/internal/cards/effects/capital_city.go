package effects

// Capital City — Land — Town (Final Fantasy, #4462):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Crystal Grotto's two mana abilities with no enters trigger: a
// colourless tap and a {1}, {T} filter into the five-colour pipe, so
// the controller picks the colour. Cycling is the shared constructor,
// activated from the hand (CR 702.29a). The land enters untapped, and
// its Town subtype gives it no mana of its own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c96ab90-155b-4bf4-acc9-65a2f0cd3189",
		Name:         "Capital City",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			painlessColorless(),
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
