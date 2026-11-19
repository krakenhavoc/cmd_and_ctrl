package effects

// Shu Elite Infantry — Creature — Human Soldier {3}{W}, 3/3
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
		OracleID:     "806b4f0c-f577-4d7f-b826-09e4479c314e",
		Name:         "Shu Elite Infantry",
		Completeness: CompletenessFull,
	})
}
