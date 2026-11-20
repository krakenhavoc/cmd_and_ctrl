package effects

// Rakdos Ickspitter — Creature — Thrull {1}{B}{R}, 1/1 (Dissension, #5598):
//
//	"{T}: This creature deals 1 damage to target creature and that
//	 creature's controller loses 1 life."
//
// disIckspitterEffect: the damage, then the life loss for the target's
// controller, both only while the target is still legal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "329f6e2e-6c37-4e98-bd45-bd252f1a2bd9",
		Name:         "Rakdos Ickspitter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target creature and that creature's controller loses 1 life.",
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  disIckspitterEffect,
		}},
	})
}
