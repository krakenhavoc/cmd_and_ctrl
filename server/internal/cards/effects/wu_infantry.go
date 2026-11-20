package effects

// Wu Infantry — Creature — Human Soldier {1}{U}, 2/1
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
		OracleID:     "a5a39eaf-104d-4dfa-b370-5d3de0d158fe",
		Name:         "Wu Infantry",
		Completeness: CompletenessFull,
	})
}
