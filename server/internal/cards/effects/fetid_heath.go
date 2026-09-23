package effects

// Fetid Heath — Land (EDHREC rank 337):
//
//	"{T}: Add {C}.
//	 {W/B}, {T}: Add {W}{W}, {W}{B}, or {B}{B}."
//
// The Guildpact filter land, Orzhov half. See rugged_prairie.go for
// the shape's history and mechanics — this is the same body with
// Orzhov colours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "42bf259d-4bb9-49c3-b4ec-223dca62f4d6",
		Name:         "Fetid Heath",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{W/B}"},
				Produced: "{W|B}{W|B}",
				Label:    "{W/B}, {T}: Add {W}{W}, {W}{B}, or {B}{B}",
			},
		},
	})
}
