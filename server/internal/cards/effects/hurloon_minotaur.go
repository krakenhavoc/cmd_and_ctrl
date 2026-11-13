package effects

// Hurloon Minotaur — Creature — Minotaur {1}{R}{R}, 2/3
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
		OracleID:     "8f1dae40-b307-446e-bbd2-86aa35813871",
		Name:         "Hurloon Minotaur",
		Completeness: CompletenessFull,
	})
}
