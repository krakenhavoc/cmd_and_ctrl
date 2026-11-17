package effects

// Air Marshal — Creature — Human Soldier {1}{U}, 2/1 (The Brothers' War,
// #5482):
//
//	"{3}: Target Soldier gains flying until end of turn."
//
// Harmattan Efreet's grant over a Soldier target, the Marshal itself
// included. A target that left or stopped being a Soldier in response
// gains nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "59b2edf0-00e4-4cd4-88ac-6dd74e9b1b14",
		Name:         "Air Marshal",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}: Target Soldier gains flying until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    ManaCost("{3}"),
			Targets: TargetCreature("target Soldier", OfCreatureType("Soldier")),
			Effect:  ebTargetCreatureGainsUntilEOT("flying", "Air Marshal — flying until end of turn"),
		}},
	})
}
