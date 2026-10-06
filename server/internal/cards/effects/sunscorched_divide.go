package effects

// Sunscorched Divide — Land:
//
//	"{1}, {T}: Add {R}{W}."
//
// The Boros filter land with no colourless ability of its own, Ferrous
// Lake's shape: a single mana ability whose cost is {1} plus the tap,
// producing the fixed pair {R}{W}. With nothing in the pool to pay the
// {1} it makes no mana, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d2b2675-19df-4f40-9e8e-196ec097b91c",
		Name:         "Sunscorched Divide",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{R}{W}",
			Label:    "{1}, {T}: Add {R}{W}",
		}},
	})
}
