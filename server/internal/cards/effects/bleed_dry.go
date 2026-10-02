package effects

// Bleed Dry — Instant {2}{B}{B}:
//
//	Target creature gets -13/-13 until end of turn. If that creature would die this turn, exile it instead.
//
// The -N/-N is applied first, then the replacement is registered on the target (ADR 0108 §1), so a creature the shrink kills is exiled.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "6c3faf4f-83c1-4098-98b8-bae15d59b0de",
		Name:         "Bleed Dry",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    shrinkFirstTargetExileIfItDies(13),
	})
}
