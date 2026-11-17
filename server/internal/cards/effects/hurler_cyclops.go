package effects

// Hurler Cyclops — Creature — Cyclops {3}{R}{R}, 5/4 (Dominaria United):
//
//	"{1}, Sacrifice another creature: This creature deals 1 damage to any
//	 target."
//
// The sacrifice is a cost, so the Cyclops itself is not a legal payment
// (SacrificeAnotherN). The damage is dealt by the Cyclops: if it has left
// by resolution, its last-known information is used (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fac57730-8fd8-4702-8b35-4ef0ed5fa41a",
		Name:         "Hurler Cyclops",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature: This creature deals 1 damage to any target.",
			Cost:    Plus(ManaCost("{1}"), SacrificeAnotherN(1, "another creature", Creature())),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  DealDamageToTheTarget(1),
		}},
	})
}
