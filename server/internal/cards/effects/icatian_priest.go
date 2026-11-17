package effects

// Icatian Priest — Creature — Human Cleric {W}, 1/1 (Fallen Empires,
// #5637):
//
//	"{1}{W}{W}: Target creature gets +1/+1 until end of turn."
//
// Any creature, the Priest itself included. No {T} in the cost, so it
// can be activated the turn the Priest arrives, and as often as it is
// paid for.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a9d29ce-e8e2-402f-b0c2-873f70418234",
		Name:         "Icatian Priest",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}{W}: Target creature gets +1/+1 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 1, 1)),
			Cost:    ManaCost("{1}{W}{W}"),
			Targets: TargetCreature("target creature"),
			Effect:  pumpAndGrantFirstTarget(1, 1, "Icatian Priest — +1/+1 until end of turn"),
		}},
	})
}
