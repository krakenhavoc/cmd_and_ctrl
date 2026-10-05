package effects

// Ferrous Lake — Land:
//
//	"{1}, {T}: Add {U}{R}."
//
// The Izzet filter land with no colourless ability of
// its own: a single mana ability whose cost is {1} plus the tap
// (ManaAbilityCost.Mana, the component Mystic Gate's filter ability
// uses), paid from the pool, producing the fixed pair {U}{R}. With
// nothing in the pool to pay the {1} it makes no mana, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "62c15af0-40e1-407d-b056-7a3d909e3fdb",
		Name:         "Ferrous Lake",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{U}{R}",
			Label:    "{1}, {T}: Add {U}{R}",
		}},
	})
}
