package effects

// Nine-Ringed Bo — Artifact {3}:
//
//	"{T}: This artifact deals 1 damage to target Spirit creature. If that
//	 creature would die this turn, exile it instead."
//
// Lava Coil's sentence on an ability: the replacement is the ability's,
// not the damage's, so the Spirit is marked whether or not the damage is
// dealt (ADR 0108 §1). "Spirit" is read off its effective subtypes, so a
// changeling counts.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f5dc3dbd-7eab-40f3-afe2-1e88a0c00538",
		Name:         "Nine-Ringed Bo",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This artifact deals 1 damage to target Spirit creature. If that creature would die this turn, exile it instead.",
			Cost:    TapCost(),
			Targets: TargetCreature("target Spirit creature", Subtype("Spirit")),
			Effect:  abilityBody(damageFirstTargetExileIfItDies(fixedAmount(1))),
		}},
	})
}
