package effects

// Ironroot Treefolk — Creature — Treefolk {4}{G}, 3/5
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
		OracleID:     "b7c0bb85-fb87-4c73-bc1b-7b4dc763c7e8",
		Name:         "Ironroot Treefolk",
		Completeness: CompletenessFull,
	})
}
