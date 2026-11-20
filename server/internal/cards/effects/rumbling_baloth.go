package effects

// Rumbling Baloth — Creature — Beast {2}{G}{G}, 4/4 (Magic 2014, #5563):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see ordinary_bear.go).
// The cost, types and body ride game.Card from Scryfall, so the Spec is the key
// and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3191c6ca-4d25-4ba3-bfe1-4aeab1295573",
		Name:         "Rumbling Baloth",
		Completeness: CompletenessFull,
	})
}
