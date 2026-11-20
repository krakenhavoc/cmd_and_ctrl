package effects

// Loxodon Convert — Creature — Phyrexian Elephant Soldier {3}{W}, 4/2 (New Phyrexia, #5575):
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
		OracleID:     "d0a80814-c69c-457c-ba19-4e1ba3fbd485",
		Name:         "Loxodon Convert",
		Completeness: CompletenessFull,
	})
}
