package effects

// Hexplate Golem — Artifact Creature — Golem {7}, 5/7 (Mirrodin Besieged,
// #5576):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// ordinary_bear.go). The cost, types and body ride game.Card from Scryfall,
// so the Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bacb9d63-5950-4bcd-a056-40c02ecf4ce8",
		Name:         "Hexplate Golem",
		Completeness: CompletenessFull,
	})
}
