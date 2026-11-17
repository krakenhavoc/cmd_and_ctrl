package effects

// Water Elemental — Creature — Elemental {3}{U}{U}, 5/4
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
		OracleID:     "c470fbe1-0538-46ec-8741-3143f5e78af8",
		Name:         "Water Elemental",
		Completeness: CompletenessFull,
	})
}
