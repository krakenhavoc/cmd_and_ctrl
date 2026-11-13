package effects

// Aeolipile — Artifact {2} (Fallen Empires, #5637):
//
//	"{1}, {T}, Sacrifice this artifact: It deals 2 damage to any target."
//
// The Aeolipile is sacrificed as part of the cost (CR 602.2) and still
// deals the damage as the ability resolves, read from its last known
// information (CR 113.7a). A target that left in response is not
// damaged (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0897adea-2759-40e8-a05a-c722473e1cf3",
		Name:         "Aeolipile",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: It deals 2 damage to any target.",
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
			Targets: TargetAny(),
			Effect:  DealDamageToTheTarget(2),
		}},
	})
}
