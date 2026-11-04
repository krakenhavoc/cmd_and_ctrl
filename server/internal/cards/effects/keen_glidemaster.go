package effects

// Keen Glidemaster — Creature — Human Soldier {1}{U}, 2/1 (Jumpstart
// Azorius deck request, #4894):
//
//	"{2}{U}: Target creature gains flying until end of turn."
//
// An instant-speed activated grant on any creature, its own included.
// A target that left or stopped being a creature in response does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "98deac9e-28ea-4d11-8d75-833bf6d66e6e",
		Name:         "Keen Glidemaster",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: Target creature gains flying until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    ManaCost("{2}{U}"),
			Targets: TargetCreature("target creature"),
			Effect:  ebTargetCreatureGainsUntilEOT("flying", "Keen Glidemaster — flying until end of turn"),
		}},
	})
}
