package effects

// Lazotep Behemoth — Creature — Zombie Hippo {4}{B}, 5/4 (War of the Spark, #5520):
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
		OracleID:     "db145875-2a40-4fe3-b8ee-bf10ae8f7cd7",
		Name:         "Lazotep Behemoth",
		Completeness: CompletenessFull,
	})
}
