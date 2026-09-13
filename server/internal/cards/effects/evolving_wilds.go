package effects

// Evolving Wilds — Land (EDHREC rank 19):
//
//	"{T}, Sacrifice this land: Search your library for a basic land
//	card, put it onto the battlefield tapped, then shuffle."
//
// The most-played nonbasic land in the format after Command Tower,
// and the cheap end of the fetch family: no life, but the land
// arrives tapped, so it costs a turn of tempo instead.
//
// Functionally identical to Terramorphic Expanse — the two are
// separate files because they are separate oracle IDs, and the
// catalog is keyed on oracle ID.
func init() {
	Register(Spec{
		OracleID:     "a75445d3-1303-4bb5-89ad-26ea93fecd48",
		Name:         "Evolving Wilds",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:  "{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Cost:   Plus(TapCost(), SacrificeThis()),
			Effect: fetchBasicTapped,
		}},
	})
}
