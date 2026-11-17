package effects

// Zuran Spellcaster — Creature — Human Wizard {2}{U}, 1/1 (Ice Age,
// #5636):
//
//	"{T}: This creature deals 1 damage to any target."
//
// Capricious Sorcerer's ping with no timing window: any target, at
// instant speed, the shared sourceDealsDamageToEachLegalTarget body. A
// target that is no longer legal as it resolves is left alone
// (CR 608.2b). The tap means summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc7b90b1-3517-4e5d-9bd8-68b4d8a259fd",
		Name:         "Zuran Spellcaster",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Cost:    TapCost(),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
