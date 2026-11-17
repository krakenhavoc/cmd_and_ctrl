package effects

// Dwarven Trader — Creature — Dwarf {R}, 1/1 (Homelands, #5635):
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
		OracleID:     "b40c5f3e-7080-4828-9215-79727c7a911f",
		Name:         "Dwarven Trader",
		Completeness: CompletenessFull,
	})
}
