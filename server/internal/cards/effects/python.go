package effects

// Python — Creature — Snake {1}{B}{B}, 3/2 (Visions, #5631):
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
		OracleID:     "d8a0c3ff-7042-4b52-a216-15e170c8094f",
		Name:         "Python",
		Completeness: CompletenessFull,
	})
}
