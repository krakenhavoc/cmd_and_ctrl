package effects

// Reduce to Ashes — Sorcery {4}{R}:
//
//	Reduce to Ashes deals 5 damage to target creature. If that creature would die this turn, exile it instead.
//
// The replacement is the spell's own effect, so it is registered on a legal target whether or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "9aab4b32-c5b3-4707-b359-9b5e3b63cd11",
		Name:         "Reduce to Ashes",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    damageFirstTargetExileIfItDies(fixedAmount(5)),
	})
}
