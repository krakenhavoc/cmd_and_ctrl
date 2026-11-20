package effects

// Lizard Warrior — Creature — Lizard Warrior {3}{R}, 4/2 (Portal, #5630):
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
		OracleID:     "123337af-f976-4956-949f-c650da812113",
		Name:         "Lizard Warrior",
		Completeness: CompletenessFull,
	})
}
