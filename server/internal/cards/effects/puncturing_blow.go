package effects

// Puncturing Blow — Sorcery {2}{R}{R}:
//
//	Puncturing Blow deals 5 damage to target creature. If that creature would die this turn, exile it instead.
//
// The replacement is the spell's own effect, so it is registered on a legal target whether or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "1128d2ab-0b6e-4912-8735-15521bc314e6",
		Name:         "Puncturing Blow",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		OnResolve:    damageFirstTargetExileIfItDies(fixedAmount(5)),
	})
}
