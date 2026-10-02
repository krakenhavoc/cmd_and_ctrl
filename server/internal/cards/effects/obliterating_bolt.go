package effects

// Obliterating Bolt — Sorcery {1}{R}:
//
//	Obliterating Bolt deals 4 damage to target creature or planeswalker. If that creature or planeswalker would die this turn, exile it instead.
//
// The replacement is the spell's own effect, so it is registered on a legal target whether or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "57fe941c-a830-4570-afe2-18f93c7a7b84",
		Name:         "Obliterating Bolt",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve:    damageFirstTargetExileIfItDies(fixedAmount(4)),
	})
}
