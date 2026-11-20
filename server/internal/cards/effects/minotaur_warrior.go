package effects

// Minotaur Warrior — Creature — Minotaur Warrior {2}{R}, 2/3 (Portal, #5630):
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
		OracleID:     "53fb9a9b-1f5d-48c9-88c0-0864c187f15e",
		Name:         "Minotaur Warrior",
		Completeness: CompletenessFull,
	})
}
