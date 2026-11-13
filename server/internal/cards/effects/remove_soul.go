package effects

// Remove Soul — Instant {1}{U} (Legends, #5639):
//
//	"Counter target creature spell."
//
// Essence Scatter's clause, on the shared counterTheTargetSpell body:
// the picker offers only creature spells, the engine refuses any other
// pick at announce, and the clause is checked again as it resolves
// (CR 608.2b). A countered spell goes to its owner's graveyard
// (CR 701.6a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b13c0f76-fbda-4911-9442-c3d7e97f1aac",
		Name:         "Remove Soul",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target creature spell", Creature()),
		OnResolve:    counterTheTargetSpell,
	})
}
