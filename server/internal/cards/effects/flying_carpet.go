package effects

// Flying Carpet — Artifact {4} (Arabian Nights, #5641):
//
//	"{2}, {T}: Target creature gains flying until end of turn."
//
// Keen Glidemaster's grant from an artifact.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "453524fb-2ffa-421f-883c-b77ae2850c52",
		Name:         "Flying Carpet",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Target creature gains flying until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  ebTargetCreatureGainsUntilEOT("flying", "Flying Carpet — flying until end of turn"),
		}},
	})
}
