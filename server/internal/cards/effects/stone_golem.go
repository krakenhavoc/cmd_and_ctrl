package effects

// Stone Golem — Artifact Creature — Golem {5}, 4/4 (Magic 2011, #5579):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7cfb32af-ac59-4459-ac9e-2762488e9180",
		Name:         "Stone Golem",
		Completeness: CompletenessFull,
	})
}
