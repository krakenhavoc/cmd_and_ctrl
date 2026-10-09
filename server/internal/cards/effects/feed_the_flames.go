package effects

// Feed the Flames — Instant {3}{R}:
//
//	Feed the Flames deals 5 damage to target creature. If that creature would die this turn, exile it instead.
//
// The replacement is the spell's own effect, so it is registered on a legal target whether or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f474d244-d9be-4580-bf62-f97660e9c1a3",
		Name:         "Feed the Flames",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		OnResolve:    damageFirstTargetExileIfItDies(fixedAmount(5)),
	})
}
