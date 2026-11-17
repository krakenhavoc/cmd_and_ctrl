package effects

// Aysen Bureaucrats — Creature — Human Advisor {1}{W}, 1/1
// (Homelands, #5635):
//
//	"{T}: Tap target creature with power 2 or less."
//
// The {T} keeps it off the turn it arrives (CR 602.5a). The power is
// read when the ability is announced and again as it resolves
// (CR 608.2b), so a creature pumped past 2 in response is not tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "786b2028-cf24-4f00-b1d6-05b94e78d617",
		Name:         "Aysen Bureaucrats",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Tap target creature with power 2 or less.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature with power 2 or less", PowerLE(2)),
			Effect:  tapTheFirstLegalTarget,
		}},
	})
}
