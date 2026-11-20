package effects

// Ogre Warrior — Creature — Ogre Warrior {3}{R}, 3/3 (Portal Second Age,
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
		OracleID:     "31426e08-bfd1-44f1-8473-5e982cb8e841",
		Name:         "Ogre Warrior",
		Completeness: CompletenessFull,
	})
}
