package effects

// Twilight Mire — Land (EDHREC rank 357):
//
//	"{T}: Add {C}.
//	 {B/G}, {T}: Add {B}{B}, {B}{G}, or {G}{G}."
//
// The Guildpact filter land, Golgari half. See rugged_prairie.go for
// the shape's history and mechanics — this is the same body with
// Golgari colours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "db623754-e078-4030-ba07-818803c348a8",
		Name:         "Twilight Mire",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{B/G}"},
				Produced: "{B|G}{B|G}",
				Label:    "{B/G}, {T}: Add {B}{B}, {B}{G}, or {G}{G}",
			},
		},
	})
}
