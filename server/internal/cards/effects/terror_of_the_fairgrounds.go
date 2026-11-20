package effects

// Terror of the Fairgrounds — Creature — Gremlin {3}{R}, 5/2 (Kaladesh, #5539):
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
		OracleID:     "75a20db5-5ce8-4e5d-a815-e69bc3dd8d4c",
		Name:         "Terror of the Fairgrounds",
		Completeness: CompletenessFull,
	})
}
