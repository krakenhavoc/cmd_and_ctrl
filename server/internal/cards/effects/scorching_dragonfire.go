package effects

// Scorching Dragonfire — Instant {1}{R}:
//
//	Scorching Dragonfire deals 3 damage to target creature or planeswalker. If that creature or planeswalker would die this turn, exile it instead.
//
// The replacement is the spell's own effect, so it is registered on a legal target whether or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "d14f313c-fea6-49c4-8197-5b74ee584a6b",
		Name:         "Scorching Dragonfire",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve:    damageFirstTargetExileIfItDies(fixedAmount(3)),
	})
}
