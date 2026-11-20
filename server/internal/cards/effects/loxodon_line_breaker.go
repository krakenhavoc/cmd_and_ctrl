package effects

// Loxodon Line Breaker — Creature — Elephant Soldier {2}{W}, 3/2
// (Core Set 2019, #5527):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "57f0f55d-4cf9-4a35-96dd-05c24c2a9a4f",
		Name:         "Loxodon Line Breaker",
		Completeness: CompletenessFull,
	})
}
