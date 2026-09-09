package effects

// Terramorphic Expanse — Land (EDHREC rank 28):
//
//	"{T}, Sacrifice this land: Search your library for a basic land
//	card, put it onto the battlefield tapped, then shuffle."
//
// Evolving Wilds under a different name. Both are in the catalog
// because both are in decks — a list that runs one very often runs
// the other as the eleventh copy.
func init() {
	Register(Spec{
		OracleID: "1bd3e453-aa21-4ee6-95c2-d6d920ee8e7a",
		Name:     "Terramorphic Expanse",
		Activated: []ActivatedAbility{{
			Label:  "{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Cost:   Plus(TapCost(), SacrificeThis()),
			Effect: fetchBasicTapped,
		}},
	})
}
