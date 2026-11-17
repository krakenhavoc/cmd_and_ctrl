package effects

// Anaba Shaman — Creature — Minotaur Shaman {3}{R}, 2/2 (Homelands, #5635):
//
//	"{R}, {T}: This creature deals 1 damage to any target."
//
// The {T} keeps it off the turn it arrives (CR 602.5a). A target that
// left in response is not damaged (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc416c17-e1f0-4344-8c21-6184cac10205",
		Name:         "Anaba Shaman",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, {T}: This creature deals 1 damage to any target.",
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Cost:    Plus(ManaCost("{R}"), TapCost()),
			Targets: TargetAny(),
			Effect:  DealDamageToTheTarget(1),
		}},
	})
}
