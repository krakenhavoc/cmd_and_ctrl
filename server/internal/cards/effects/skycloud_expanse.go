package effects

// Skycloud Expanse — Land (EDHREC rank 345):
//
//	"{1}, {T}: Add {W}{U}."
//
// The Odyssey "tap-and-a-tax" dual, Azorius half. See
// darkwater_catacombs.go for the shape's mechanics — this is the
// same body with Azorius colours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "76f335d0-7f71-4b1a-b60d-73de954cbe2c",
		Name:         "Skycloud Expanse",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W}{U}",
				Label:    "{1}, {T}: Add {W}{U}",
			},
		},
	})
}
