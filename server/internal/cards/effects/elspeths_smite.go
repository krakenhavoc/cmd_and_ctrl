package effects

// Elspeth's Smite — Instant {W}:
//
//	Elspeth's Smite deals 3 damage to target attacking or blocking creature. If that creature would die this turn, exile it instead.
//
// The replacement is the spell's own effect, so it is registered on a legal target whether or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "3f404fe4-4335-4dcc-ba90-78246c4b880b",
		Name:         "Elspeth's Smite",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve:    damageFirstTargetExileIfItDies(fixedAmount(3)),
	})
}
