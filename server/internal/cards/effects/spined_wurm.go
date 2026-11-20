package effects

// Spined Wurm — Creature — Wurm {4}{G}, 5/4 (Portal, #5630):
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
		OracleID:     "05425f2f-7228-4bf5-8fe1-6fe99107e8e0",
		Name:         "Spined Wurm",
		Completeness: CompletenessFull,
	})
}
