package effects

// Magma Spray — Instant {R}:
//
//	Magma Spray deals 2 damage to target creature. If that creature would die this turn, exile it instead.
//
// The replacement is the spell's own effect, so it is registered on a legal target whether or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "fe16f1ab-58b4-4452-abe4-cbd9addd348f",
		Name:         "Magma Spray",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		OnResolve:    damageFirstTargetExileIfItDies(fixedAmount(2)),
	})
}
