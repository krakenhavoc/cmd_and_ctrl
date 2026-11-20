package effects

// Omega Myr — Artifact Creature — Myr {2}, 1/2 (Mirrodin, #5606):
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
		OracleID:     "fd1b237a-7132-4875-b738-8ec65b85ea28",
		Name:         "Omega Myr",
		Completeness: CompletenessFull,
	})
}
