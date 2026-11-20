package effects

// Rowan Treefolk — Creature — Treefolk {3}{G}, 3/4 (Portal, #5630):
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
		OracleID:     "ae37dd07-2926-4801-b89d-3ffad9f0f575",
		Name:         "Rowan Treefolk",
		Completeness: CompletenessFull,
	})
}
