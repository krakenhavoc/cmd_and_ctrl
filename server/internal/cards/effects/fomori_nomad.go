package effects

// Fomori Nomad — Creature — Nomad Giant {4}{R}, 4/4
// (Future Sight, #5594):
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
		OracleID:     "bfcaa0a4-2fc5-4e27-8b79-802e365c71f2",
		Name:         "Fomori Nomad",
		Completeness: CompletenessFull,
	})
}
