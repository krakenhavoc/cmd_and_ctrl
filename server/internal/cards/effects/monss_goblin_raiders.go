package effects

// Mons's Goblin Raiders — Creature — Goblin {R}, 1/1
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
		OracleID:     "a37159df-f6d7-4db6-85de-0ea77f425993",
		Name:         "Mons's Goblin Raiders",
		Completeness: CompletenessFull,
	})
}
