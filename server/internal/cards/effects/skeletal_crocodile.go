package effects

// Skeletal Crocodile — Creature — Crocodile Skeleton {3}{B}, 5/1 (Portal, #5630):
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
		OracleID:     "90f9440a-0bf8-46eb-9cfa-16fb88a217d8",
		Name:         "Skeletal Crocodile",
		Completeness: CompletenessFull,
	})
}
