package effects

// Goblin Piker — Creature — Goblin Warrior {1}{R}, 2/1 (Portal Second Age,
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
		OracleID:     "50608184-90d3-43d2-a221-deb186c78323",
		Name:         "Goblin Piker",
		Completeness: CompletenessFull,
	})
}
