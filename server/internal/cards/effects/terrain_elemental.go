package effects

// Terrain Elemental — Creature — Elemental {1}{G}, 3/2 (Kaladesh, #5539):
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
		OracleID:     "3a9b9fe0-1846-41fc-b7c3-ed9abef27a6a",
		Name:         "Terrain Elemental",
		Completeness: CompletenessFull,
	})
}
