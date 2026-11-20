package effects

// Barbarian Horde — Creature — Human Barbarian Soldier {3}{R}, 3/3
// (Portal Three Kingdoms, #5622):
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
		OracleID:     "eca9cb24-4dc8-48eb-b055-5fdf7c67bd43",
		Name:         "Barbarian Horde",
		Completeness: CompletenessFull,
	})
}
