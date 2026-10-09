package effects

// Incendiary Flow — Sorcery {1}{R}:
//
//	Incendiary Flow deals 3 damage to any target. If a creature dealt damage this way would die this turn, exile it instead.
//
// "A creature dealt damage this way" is registered from the damage's continuation, on a creature that was dealt more than 0 damage (ADR 0108 §1 decision 3).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "91b2ffe8-155d-4b9f-82dd-868cc895856b",
		Name:         "Incendiary Flow",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve:    damageAnyTargetExileIfDealtDies(fixedAmount(3), false),
	})
}
