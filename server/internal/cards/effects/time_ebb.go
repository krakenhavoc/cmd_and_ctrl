package effects

// Time Ebb — Sorcery {2}{U} (Portal, #5630):
//
//	"Put target creature on top of its owner's library."
//
// Temporal Spring's body, on a creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30cc8f7b-3c28-40f5-8f8f-157e8212280b",
		Name:         "Time Ebb",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    putTargetOnTopOfOwnersLibrary,
	})
}
