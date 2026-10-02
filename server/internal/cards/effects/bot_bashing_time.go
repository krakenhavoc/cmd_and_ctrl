package effects

// Bot Bashing Time — Sorcery {3}{R}:
//
//	Bot Bashing Time deals 6 damage to target creature. If that creature would die this turn, exile it instead.
//
// The replacement is the spell's own effect, so it is registered on a legal target whether or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "14079f05-fbc1-401c-9008-da678656b4c6",
		Name:         "Bot Bashing Time",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    damageFirstTargetExileIfItDies(fixedAmount(6)),
	})
}
