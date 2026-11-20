package effects

// Talas Merchant — Creature — Human Pirate {1}{U}, 1/3 (Portal Second Age,
// #5625):
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
		OracleID:     "6469f61d-1b4c-4813-bb4a-99f6e6388d7b",
		Name:         "Talas Merchant",
		Completeness: CompletenessFull,
	})
}
