package effects

// Pinnacle of Rage — Sorcery {4}{R}{R} (Born of the Gods,
// #5559):
//
//	"Pinnacle of Rage deals 3 damage to each of two targets."
//
// Furious Reprisal at 3: exactly two distinct "any target" picks, and
// a target that became illegal takes nothing while the other is still
// dealt its 3.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ca999d7-48c9-4517-8e63-82ddccdb430e",
		Name:         "Pinnacle of Rage",
		Completeness: CompletenessFull,
		Targets:      TargetAny().WithCount(2, 2),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve:    dealDamageToEachLegalTarget(3),
	})
}
