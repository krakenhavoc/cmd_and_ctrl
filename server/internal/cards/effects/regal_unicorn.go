package effects

// Regal Unicorn — Creature — Unicorn {2}{W}, 2/3 (Portal, #5630):
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
		OracleID:     "2b856a43-ca6b-4f08-b610-9794ee6d7fcf",
		Name:         "Regal Unicorn",
		Completeness: CompletenessFull,
	})
}
