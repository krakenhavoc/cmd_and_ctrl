package effects

// Goblin Cavaliers — Creature — Goblin {2}{R}, 3/2 (Portal Second Age,
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
		OracleID:     "be77e3f8-76a3-4d03-a8be-ad4b0b72f929",
		Name:         "Goblin Cavaliers",
		Completeness: CompletenessFull,
	})
}
