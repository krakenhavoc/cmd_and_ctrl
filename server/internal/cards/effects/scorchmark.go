package effects

// Scorchmark — Instant {1}{R}:
//
//	Scorchmark deals 2 damage to target creature. If that creature would die this turn, exile it instead.
//
// The replacement is the spell's own effect, so it is registered on a legal target whether or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "8dc1148f-c6bc-469c-8d1a-7e3efd2de7e2",
		Name:         "Scorchmark",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		Targets:      TargetCreature("target creature"),
		OnResolve:    damageFirstTargetExileIfItDies(fixedAmount(2)),
	})
}
