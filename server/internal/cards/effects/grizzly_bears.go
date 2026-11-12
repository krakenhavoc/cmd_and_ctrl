package effects

// Grizzly Bears — Creature — Bear {1}{G}, 2/2
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
		OracleID:     "14c8f55d-d177-4c25-a931-ebeb9e6062a0",
		Name:         "Grizzly Bears",
		Completeness: CompletenessFull,
	})
}
