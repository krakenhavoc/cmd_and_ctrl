package effects

// Zephyr Charge — Enchantment {1}{U} (Magic 2014, #5563):
//
//	"{1}{U}: Target creature gains flying until end of turn."
//
// Harmattan Efreet's grant at {1}{U}, from an enchantment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "576128ba-28aa-4ef9-9e72-8ca13f42c9b4",
		Name:         "Zephyr Charge",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}: Target creature gains flying until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    ManaCost("{1}{U}"),
			Targets: TargetCreature("target creature"),
			Effect:  ebTargetCreatureGainsUntilEOT("flying", "Zephyr Charge — flying until end of turn"),
		}},
	})
}
