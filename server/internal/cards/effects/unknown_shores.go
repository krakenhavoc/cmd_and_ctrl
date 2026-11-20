package effects

// Unknown Shores — Land (Theros, #5561):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color."
//
// Painted Bluffs' two mana abilities: colorless for free, or any color
// for {1}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "82c5ec8e-27be-475a-9921-ad61209fd022",
		Name:         "Unknown Shores",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			painlessColorless(),
			{
				Cost:     ManaAbilityCost{Mana: "{1}", Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
	})
}
