package effects

// Goblin Roughrider — Creature — Goblin Knight {2}{R}, 3/2 (Worldwake, #5582):
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
		OracleID:     "a920d37e-e172-41f7-bfbf-c18821dcbed7",
		Name:         "Goblin Roughrider",
		Completeness: CompletenessFull,
	})
}
