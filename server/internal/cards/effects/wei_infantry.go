package effects

// Wei Infantry — Creature — Human Soldier {1}{B}, 2/1
// (Portal Three Kingdoms, #5622):
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
		OracleID:     "41283a82-ce3f-41a6-995e-bc30d8ab5cf9",
		Name:         "Wei Infantry",
		Completeness: CompletenessFull,
	})
}
