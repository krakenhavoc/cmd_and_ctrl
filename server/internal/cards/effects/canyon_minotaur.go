package effects

// Canyon Minotaur — Creature — Minotaur Warrior {3}{R}, 3/3 (Conflux,
// #5587):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it. The cost, types
// and body ride game.Card from Scryfall, so the Spec is the key and the
// name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c7163e6-f5a1-45a5-88c2-19dd9ef0a587",
		Name:         "Canyon Minotaur",
		Completeness: CompletenessFull,
	})
}
