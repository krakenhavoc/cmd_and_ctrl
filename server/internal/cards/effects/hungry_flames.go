package effects

// Hungry Flames — Instant {2}{R} (Aether Revolt,
// #5536):
//
//	"Hungry Flames deals 3 damage to target creature and 2 damage to target
//	 player or planeswalker."
//
// Two clauses, Arc Trail's shape (damageEachClauseTarget): one damage
// instance, and a clause whose target left is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2c00788-14e7-4ab5-8c71-2a176621a177",
		Name:         "Hungry Flames",
		Completeness: CompletenessFull,
		Targets:      Clauses(TargetCreature("target creature"), targetPlayerOrPlaneswalker()),
		Purpose:      ForTargets(DamageToTarget(0, 3), DamageToTarget(1, 2)),
		OnResolve:    damageEachClauseTarget(3, 2),
	})
}
