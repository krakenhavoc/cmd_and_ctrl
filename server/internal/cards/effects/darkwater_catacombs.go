package effects

// Darkwater Catacombs — Land (EDHREC rank 323):
//
//	"{1}, {T}: Add {U}{B}."
//
// An Odyssey "tap-and-a-tax" dual: no colourless option at all, one
// ability, {1} to filter into a fixed two colours. Mechanically the
// Signet cycle's mana component (`ManaAbilityCost.Mana`, closed by
// #356) on a land instead of an artifact — the engine has no auto-tap
// into it, same as every card built on that field.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4869a530-757f-4364-8d8e-4dc8001f433c",
		Name:         "Darkwater Catacombs",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{U}{B}",
				Label:    "{1}, {T}: Add {U}{B}",
			},
		},
	})
}
