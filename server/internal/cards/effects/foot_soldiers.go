package effects

// Foot Soldiers — Creature — Human Soldier {3}{W}, 2/4 (Portal, #5630):
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
		OracleID:     "a768ba13-4d1c-4dce-a4a6-86a39c069c3f",
		Name:         "Foot Soldiers",
		Completeness: CompletenessFull,
	})
}
