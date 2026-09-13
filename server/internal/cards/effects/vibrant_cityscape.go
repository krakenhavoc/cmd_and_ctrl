package effects

// Vibrant Cityscape — Land (EDHREC rank 1454):
//
//	"{T}, Sacrifice this land: Search your library for a basic land
//	 card, put it onto the battlefield tapped, then shuffle."
//
// Evolving Wilds with a new name: the searcher picks the basic, it
// arrives tapped, the library shuffles. Its own file because the
// catalog is keyed on oracle ID, exactly as Terramorphic Expanse is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a7f3e1f-6798-4644-b64c-7765f81f0938",
		Name:         "Vibrant Cityscape",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:  "{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Cost:   Plus(TapCost(), SacrificeThis()),
			Effect: fetchBasicTapped,
		}},
	})
}
