package effects

// Flooded Grove — Land (EDHREC rank 302):
//
//	"{T}: Add {C}.
//	 {G/U}, {T}: Add {G}{G}, {G}{U}, or {U}{U}."
//
// The Guildpact filter land, Simic half. See rugged_prairie.go for
// the shape's history and mechanics — this is the same body with
// Simic colours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc974eb4-72b9-4213-887b-8ee684b93420",
		Name:         "Flooded Grove",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{G/U}"},
				Produced: "{G|U}{G|U}",
				Label:    "{G/U}, {T}: Add {G}{G}, {G}{U}, or {U}{U}",
			},
		},
	})
}
