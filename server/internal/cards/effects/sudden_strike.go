package effects

// Sudden Strike — Instant {1}{W} (Marvel Super Heroes Commander,
// #4455): "Destroy target attacking or blocking creature." The target
// clause is Elspeth's Smite's AttackingOrBlocking; the body is the shared
// destroy-the-target step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8147a759-e387-48f6-9c1e-bb39622ed553",
		Name:         "Sudden Strike",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
