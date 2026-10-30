package effects

// Repulsor Rays — Sorcery {R} (Marvel Super Heroes Commander, #4455):
// "Repulsor Rays deals 3 damage to target creature." Concentrated
// Fire's damage body at three damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "053a90b8-61b1-4390-afc0-c5eaef0138a9",
		Name:         "Repulsor Rays",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve:    damageToFirstTarget(3),
	})
}
