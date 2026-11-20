package effects

// Independent Troops — Creature — Human Soldier {1}{R}, 2/1
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
		OracleID:     "b9a8adf0-60f1-491c-8bbf-0b341fd23fab",
		Name:         "Independent Troops",
		Completeness: CompletenessFull,
	})
}
