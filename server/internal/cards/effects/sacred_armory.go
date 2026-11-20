package effects

// Sacred Armory — Artifact {2} (Magic 2015, #5556):
//
//	"{2}: Target creature gets +1/+0 until end of turn."
//
// Captive Flame's pump at {2} (neoB07TargetGetsUntilEOT): no {T}, so it
// stacks with each activation, and a target that left is skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "37d74df9-cf87-4b82-842a-3941064078fb",
		Name:         "Sacred Armory",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}: Target creature gets +1/+0 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 1, 0)),
			Cost:    ManaCost("{2}"),
			Targets: TargetCreature("target creature"),
			Effect:  neoB07TargetGetsUntilEOT("Sacred Armory — +1/+0 until end of turn", 1, 0),
		}},
	})
}
