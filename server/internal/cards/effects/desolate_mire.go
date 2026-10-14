package effects

// Desolate Mire — Land (EDHREC top 1,000, #3654):
//
//	"{1}, {T}: Add {W}{B}."
//
// A filter land: no colourless option, one ability, {1} to filter into
// a fixed two colours. The Signet cycle's mana component
// (ManaAbilityCost.Mana) on a land, as Darkwater Catacombs is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3edf9201-265f-4cd9-b27b-8073bf1a4cf2",
		Name:         "Desolate Mire",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{W}{B}",
			Label:    "{1}, {T}: Add {W}{B}",
		}},
	})
}
