package effects

// Bloodlust Inciter — Creature — Human Warrior {R}, 1/1 (Foundations
// Jumpstart):
//
//	"{T}: Target creature gains haste until end of turn."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "826b83ba-0fe5-4f1d-91c8-799a1100d1e6",
		Name:         "Bloodlust Inciter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature gains haste until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  ebTargetCreatureGainsUntilEOT("haste", "Bloodlust Inciter — haste until end of turn"),
		}},
	})
}
