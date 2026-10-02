package effects

// Disintegrate — Sorcery {X}{R}:
//
//	"Disintegrate deals X damage to any target. If it's a creature, it
//	 can't be regenerated this turn, and if it would die this turn, exile
//	 it instead."
//
// Both riders are effects of the spell, not of the damage. The
// Disintegrate ruling: "The 'can't regenerate' is an effect of
// Disintegrate and not an effect of the damage. It works even if the
// damage is prevented or redirected." So a creature target is marked
// whatever it was dealt (ADR 0108 §1, §2).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "92d6af2f-728e-4e41-87cb-5c90878a2f2f",
		Name:         "Disintegrate",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		OnResolve:    damageThenIfCreatureNoRegenExileIfDies(xAmount()),
	})
}
