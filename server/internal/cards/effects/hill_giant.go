package effects

// Hill Giant — Creature — Giant {3}{R}, 3/3
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
		OracleID:     "342199e0-15b6-4824-83da-25caef2592b3",
		Name:         "Hill Giant",
		Completeness: CompletenessFull,
	})
}
