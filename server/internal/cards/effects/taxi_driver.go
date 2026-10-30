package effects

// Taxi Driver — Creature — Human Pilot {1}{R}, 3/1 (Marvel Super Heroes
// Commander, #4455):
//
//	"{1}, {T}: Target creature gains haste until end of turn."
//
// Hanweir Battlements' haste ability, paid with {1} instead of {R}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a17a92f7-e95c-4d0e-a43e-7e6b41a59139",
		Name:         "Taxi Driver",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Target creature gains haste until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  ebTargetCreatureGainsUntilEOT("haste", "Taxi Driver — haste until end of turn"),
		}},
	})
}
