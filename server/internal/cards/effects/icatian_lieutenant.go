package effects

// Icatian Lieutenant — Creature — Human Soldier {W}{W}, 1/2 (Fallen
// Empires, #5637):
//
//	"{1}{W}: Target Soldier creature gets +1/+0 until end of turn."
//
// The target must be a Soldier as the ability is activated and again
// as it resolves (CR 608.2b); the Lieutenant is a Soldier, so it may
// pump itself. No {T} in the cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "51e9af95-745b-45b9-b870-7f768578fa4f",
		Name:         "Icatian Lieutenant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: Target Soldier creature gets +1/+0 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 1, 0)),
			Cost:    ManaCost("{1}{W}"),
			Targets: TargetCreature("target Soldier creature", HasSubtype("Soldier")),
			Effect:  pumpAndGrantFirstTarget(1, 0, "Icatian Lieutenant — +1/+0 until end of turn"),
		}},
	})
}
