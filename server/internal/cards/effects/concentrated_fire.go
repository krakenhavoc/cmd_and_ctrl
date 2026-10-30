package effects

// Concentrated Fire — Instant {3}{R} (Marvel Super Heroes Commander, #4455):
// "Concentrated Fire deals 5 damage to target creature." Lightning Axe's
// damage body without the additional cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2fcb2342-202d-4baa-82ad-66d782142264",
		Name:         "Concentrated Fire",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		OnResolve:    damageToFirstTarget(5),
	})
}
