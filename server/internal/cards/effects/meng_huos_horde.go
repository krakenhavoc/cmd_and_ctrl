package effects

// Meng Huo's Horde — Creature — Human Soldier {4}{G}, 4/5
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
		OracleID:     "27e6de72-68c4-4154-807f-5944fe97c92b",
		Name:         "Meng Huo's Horde",
		Completeness: CompletenessFull,
	})
}
