package effects

// Deadly Riposte — Instant {1}{W} (Foundations, #4465):
//
//	"Deadly Riposte deals 3 damage to target tapped creature and you gain
//	 2 life."
//
// Searing Blood's shape: the damage is the first step and the life gain
// the second, so the gain waits for any prompt the damage asks (CR
// 608.2c). The target must be tapped as the spell is cast and again as
// it resolves; a creature that has untapped by then is an illegal target,
// and a spell whose only target is illegal doesn't resolve, life gain and
// all (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e341478e-c3f8-45d9-9430-abe3a58fa2fc",
		Name:         "Deadly Riposte",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target tapped creature", b751Tapped()),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve: ResolveSteps(
			damageTheClauseTarget(0, 3),
			GainLife{Amount: 2},
		),
	})
}
