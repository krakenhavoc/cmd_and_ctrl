package effects

// Caller of Gales — Creature — Merfolk Wizard {U}, 1/1 (Zendikar, #5583):
//
//	"{1}{U}, {T}: Target creature gains flying until end of turn."
//
// Flying Carpet's row at {1}{U}: the tap is part of the cost, so
// summoning sickness stops it (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "86fd8f95-0f1c-42c0-a63e-567c59ce89c4",
		Name:         "Caller of Gales",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, {T}: Target creature gains flying until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    Plus(ManaCost("{1}{U}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  ebTargetCreatureGainsUntilEOT("flying", "Caller of Gales — flying until end of turn"),
		}},
	})
}
