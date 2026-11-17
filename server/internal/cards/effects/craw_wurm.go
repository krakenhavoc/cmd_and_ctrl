package effects

// Craw Wurm — Creature — Wurm {4}{G}{G}, 6/4
// (Limited Edition Alpha, #5642):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// ordinary_bear.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a462a69-3e42-41de-a3aa-a488d9f38d69",
		Name:         "Craw Wurm",
		Completeness: CompletenessFull,
	})
}
