package effects

// Trained Jackal — Creature — Jackal {G}, 1/2 (Portal Three Kingdoms, #5622):
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
		OracleID:     "99545320-b6ad-4caf-b29a-47f637395ef7",
		Name:         "Trained Jackal",
		Completeness: CompletenessFull,
	})
}
