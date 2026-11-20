package effects

// Nessian Courser — Creature — Centaur Warrior {2}{G}, 3/3
// (Future Sight, #5594):
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
		OracleID:     "e876d1fc-3acd-41a3-a34b-2bfa83204393",
		Name:         "Nessian Courser",
		Completeness: CompletenessFull,
	})
}
