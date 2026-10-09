package effects

// Flame-Blessed Bolt — Instant {R}:
//
//	Flame-Blessed Bolt deals 2 damage to target creature or planeswalker. If that creature or planeswalker would die this turn, exile it instead.
//
// The replacement is the spell's own effect, so it is registered on a legal target whether or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "6a521eed-0965-4fce-8a11-190dc2863da8",
		Name:         "Flame-Blessed Bolt",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		OnResolve:    damageFirstTargetExileIfItDies(fixedAmount(2)),
	})
}
