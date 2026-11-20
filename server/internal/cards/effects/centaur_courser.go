package effects

// Centaur Courser — Creature — Centaur Warrior {2}{G}, 3/3 (Magic 2010, #5585):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f5bf099-2e01-4e1c-9ebf-0ce0ac66939e",
		Name:         "Centaur Courser",
		Completeness: CompletenessFull,
	})
}
