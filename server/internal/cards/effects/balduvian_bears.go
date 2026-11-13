package effects

// Balduvian Bears — Creature — Bear {1}{G}, 2/2 (Ice Age, #5636):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "32c10a3e-85b0-49af-9b23-81d0b7e407df",
		Name:         "Balduvian Bears",
		Completeness: CompletenessFull,
	})
}
