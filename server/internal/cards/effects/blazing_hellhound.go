package effects

// Blazing Hellhound — Creature — Elemental Dog {2}{B}{R}, 4/3 (Magic Origins, #5549):
//
//	"{1}, Sacrifice another creature: This creature deals 1 damage to any target."
//
// Hurler Cyclops's ability: the sacrifice is another creature, paid as a
// cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "598a7410-35e8-4d62-9bf1-d44ee7717040",
		Name:         "Blazing Hellhound",
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
