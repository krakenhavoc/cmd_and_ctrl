package effects

// Barony Vampire — Creature — Vampire {2}{B}, 3/2 (Magic 2011, #5579):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e886756a-298d-4f79-8699-8bfc9345352e",
		Name:         "Barony Vampire",
		Completeness: CompletenessFull,
	})
}
