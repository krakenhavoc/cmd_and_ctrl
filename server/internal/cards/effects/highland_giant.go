package effects

// Highland Giant — Creature — Giant {2}{R}{R}, 3/4 (Portal, #5630):
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
		OracleID:     "2f3878b8-61af-4276-b7ab-11b70e44eb62",
		Name:         "Highland Giant",
		Completeness: CompletenessFull,
	})
}
