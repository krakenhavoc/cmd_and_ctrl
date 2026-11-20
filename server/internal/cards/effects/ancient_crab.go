package effects

// Ancient Crab — Creature — Crab {1}{U}{U}, 1/5 (Oath of the Gatewatch, #5545):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6bab6d0a-cd06-45e3-b780-d5354ee2a032",
		Name:         "Ancient Crab",
		Completeness: CompletenessFull,
	})
}
