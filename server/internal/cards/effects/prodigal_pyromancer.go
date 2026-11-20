package effects

// Prodigal Pyromancer — Creature — Human Wizard {2}{R}, 1/1 (Planar Chaos, #5595):
//
//	"{T}: This creature deals 1 damage to any target."
//
// Anaba Shaman's ping without the mana: the {T} keeps it off the turn it
// arrives (CR 302.6), and a target that left in response is not damaged
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "29b3cb3e-aeb1-43eb-8095-9f7881aa34bd",
		Name:         "Prodigal Pyromancer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Cost:    TapCost(),
			Targets: TargetAny(),
			Effect:  DealDamageToTheTarget(1),
		}},
	})
}
