package effects

// Goblin Hero — Creature — Goblin {2}{R}, 2/2
// (The Dark, #5638):
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
		OracleID:     "ee969637-a20e-4163-97c0-9fd5cb17b741",
		Name:         "Goblin Hero",
		Completeness: CompletenessFull,
	})
}
