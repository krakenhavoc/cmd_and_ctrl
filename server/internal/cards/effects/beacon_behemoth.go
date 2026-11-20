package effects

// Beacon Behemoth — Creature — Beast {3}{G}{G}, 5/3 (Conflux, #5587):
//
//	"{1}: Target creature with power 5 or greater gains vigilance until
//	 end of turn."
//
// Spearbreaker Behemoth's row with vigilance. The power is the target's
// current power, checked as the ability is activated and again as it
// resolves (CR 608.2b); it can target the Behemoth itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "25571850-d5c6-460f-b212-9fc9aa705277",
		Name:         "Beacon Behemoth",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}: Target creature with power 5 or greater gains vigilance until end of turn.",
			Cost:    ManaCost("{1}"),
			Targets: TargetCreature("target creature with power 5 or greater", PowerGE(5)),
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "vigilance")),
			Effect:  ebTargetCreatureGainsUntilEOT("vigilance", "Beacon Behemoth — vigilance"),
		}},
	})
}
