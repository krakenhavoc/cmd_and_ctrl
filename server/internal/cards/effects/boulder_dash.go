package effects

// Boulder Dash — Sorcery {1}{R} (Lorwyn Eclipsed, #4458):
//
//	"Boulder Dash deals 2 damage to any target and 1 damage to any other
//	 target."
//
// Arc Trail's text and shape, and its body (damageEachClauseTarget): two
// clauses, slot 0 takes 2 and slot 1 takes 1, and slot 1 is Distinct
// ("any other target"). Each slot is re-checked on its own at resolution
// (CR 608.2b), so if one target has left the other is still dealt its
// damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30a9ad0e-1b23-41ce-bacc-3b27f78ef628",
		Name:         "Boulder Dash",
		Completeness: CompletenessFull,
		Targets:      Clauses(TargetAny(), Distinct(TargetAny())),
		Purpose:      ForTargets(DamageToTarget(0, 2), DamageToTarget(1, 1)),
		OnResolve:    damageEachClauseTarget(2, 1),
	})
}
