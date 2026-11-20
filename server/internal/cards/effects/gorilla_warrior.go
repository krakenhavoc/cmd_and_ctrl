package effects

// Gorilla Warrior — Creature — Ape Warrior {2}{G}, 3/2 (Portal, #5630):
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
		OracleID:     "c0e6ae0f-6cf7-48ca-acb8-73d4c38b9005",
		Name:         "Gorilla Warrior",
		Completeness: CompletenessFull,
	})
}
