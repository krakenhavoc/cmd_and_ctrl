package effects

// Curio Vendor — Creature — Vedalken {1}{U}, 2/1 (Kaladesh, #5539):
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
		OracleID:     "7db99e21-0faf-408d-a770-189726179c66",
		Name:         "Curio Vendor",
		Completeness: CompletenessFull,
	})
}
