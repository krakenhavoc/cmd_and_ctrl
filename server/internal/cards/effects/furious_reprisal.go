package effects

// Furious Reprisal — Sorcery {3}{R} (Foundations Jumpstart):
//
//	"Furious Reprisal deals 2 damage to each of two targets."
//
// Volcanic Salvo's damage to each target, over exactly two distinct
// "any target" picks. A target that became illegal takes nothing and the
// other still does (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "38e1ab26-95f9-43dd-9bda-13c7d8ba47ea",
		Name:         "Furious Reprisal",
		Completeness: CompletenessFull,
		Targets:      TargetAny().WithCount(2, 2),
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		OnResolve:    dealDamageToEachLegalTarget(2),
	})
}
