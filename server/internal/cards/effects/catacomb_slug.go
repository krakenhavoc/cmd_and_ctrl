package effects

// Catacomb Slug — Creature — Slug {4}{B}, 2/6 (Return to Ravnica, #5566):
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
		OracleID:     "1b0b5264-2810-4ac5-8b2e-8abbf4016287",
		Name:         "Catacomb Slug",
		Completeness: CompletenessFull,
	})
}
